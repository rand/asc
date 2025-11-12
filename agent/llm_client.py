"""
LLM Client Abstraction - Unified interface for multiple LLM providers.

Supports Claude (Anthropic), Gemini (Google), and OpenAI (GPT-4, Codex).
"""

import os
import time
import logging
from abc import ABC, abstractmethod
from typing import Dict, Optional, Any
from dataclasses import dataclass


@dataclass
class CompletionResult:
    """Result from an LLM completion request."""
    content: str
    tokens_used: int
    cost_usd: float
    model: str
    finish_reason: str


class LLMClient(ABC):
    """Abstract base class for LLM clients."""
    
    def __init__(self, model: str):
        self.model = model
        self.logger = logging.getLogger(f"llm.{model}")
        self.total_tokens = 0
        self.total_cost = 0.0
        self.request_count = 0
    
    @abstractmethod
    def complete(
        self,
        prompt: str,
        system_prompt: Optional[str] = None,
        max_tokens: int = 4096,
        temperature: float = 0.7
    ) -> CompletionResult:
        """
        Generate a completion from the LLM.
        
        Args:
            prompt: The user prompt
            system_prompt: Optional system prompt
            max_tokens: Maximum tokens to generate
            temperature: Sampling temperature (0.0 to 1.0)
            
        Returns:
            CompletionResult with the generated text and metadata
        """
        pass
    
    def _retry_with_backoff(self, func, max_retries: int = 3):
        """Execute a function with exponential backoff retry logic."""
        for attempt in range(max_retries):
            try:
                return func()
            except Exception as e:
                if attempt == max_retries - 1:
                    raise
                
                wait_time = 2 ** attempt
                self.logger.warning(
                    f"Request failed (attempt {attempt + 1}/{max_retries}): {e}. "
                    f"Retrying in {wait_time}s..."
                )
                time.sleep(wait_time)
    
    def get_stats(self) -> Dict[str, Any]:
        """Get usage statistics."""
        return {
            "model": self.model,
            "total_tokens": self.total_tokens,
            "total_cost_usd": round(self.total_cost, 4),
            "request_count": self.request_count,
            "avg_tokens_per_request": (
                self.total_tokens // self.request_count if self.request_count > 0 else 0
            )
        }


class ClaudeClient(LLMClient):
    """Client for Anthropic's Claude models."""
    
    def __init__(self, model: str = "claude-3-5-sonnet-20241022"):
        super().__init__(model)
        self.api_key = os.getenv("CLAUDE_API_KEY")
        if not self.api_key:
            raise ValueError("CLAUDE_API_KEY environment variable not set")
        
        try:
            from anthropic import Anthropic
            self.client = Anthropic(api_key=self.api_key)
        except ImportError:
            raise ImportError("anthropic package not installed. Run: pip install anthropic")
    
    def complete(
        self,
        prompt: str,
        system_prompt: Optional[str] = None,
        max_tokens: int = 4096,
        temperature: float = 0.7
    ) -> CompletionResult:
        """Generate completion using Claude."""
        
        def _make_request():
            messages = [{"role": "user", "content": prompt}]
            
            kwargs = {
                "model": self.model,
                "max_tokens": max_tokens,
                "temperature": temperature,
                "messages": messages
            }
            
            if system_prompt:
                kwargs["system"] = system_prompt
            
            response = self.client.messages.create(**kwargs)
            
            # Extract content
            content = ""
            for block in response.content:
                if hasattr(block, "text"):
                    content += block.text
            
            # Calculate tokens and cost
            input_tokens = response.usage.input_tokens
            output_tokens = response.usage.output_tokens
            total_tokens = input_tokens + output_tokens
            
            # Claude pricing (approximate, as of 2024)
            # Sonnet: $3/MTok input, $15/MTok output
            cost = (input_tokens * 3.0 / 1_000_000) + (output_tokens * 15.0 / 1_000_000)
            
            self.total_tokens += total_tokens
            self.total_cost += cost
            self.request_count += 1
            
            return CompletionResult(
                content=content,
                tokens_used=total_tokens,
                cost_usd=cost,
                model=self.model,
                finish_reason=response.stop_reason or "complete"
            )
        
        try:
            result = self._retry_with_backoff(_make_request)
            self.logger.info(
                f"Completion successful: {result.tokens_used} tokens, "
                f"${result.cost_usd:.4f}"
            )
            return result
        except Exception as e:
            self.logger.error(f"Claude API error: {e}", exc_info=True)
            raise


class GeminiClient(LLMClient):
    """Client for Google's Gemini models."""
    
    def __init__(self, model: str = "gemini-1.5-pro"):
        super().__init__(model)
        self.api_key = os.getenv("GOOGLE_API_KEY")
        if not self.api_key:
            raise ValueError("GOOGLE_API_KEY environment variable not set")
        
        try:
            import google.generativeai as genai
            genai.configure(api_key=self.api_key)
            self.client = genai.GenerativeModel(model)
            self.genai = genai
        except ImportError:
            raise ImportError(
                "google-generativeai package not installed. "
                "Run: pip install google-generativeai"
            )
        
        # Rate limiting
        self.last_request_time = 0
        self.min_request_interval = 1.0  # 1 second between requests
    
    def complete(
        self,
        prompt: str,
        system_prompt: Optional[str] = None,
        max_tokens: int = 4096,
        temperature: float = 0.7
    ) -> CompletionResult:
        """Generate completion using Gemini."""
        
        # Rate limiting
        elapsed = time.time() - self.last_request_time
        if elapsed < self.min_request_interval:
            time.sleep(self.min_request_interval - elapsed)
        
        def _make_request():
            # Combine system prompt and user prompt
            full_prompt = prompt
            if system_prompt:
                full_prompt = f"{system_prompt}\n\n{prompt}"
            
            generation_config = self.genai.GenerationConfig(
                max_output_tokens=max_tokens,
                temperature=temperature
            )
            
            response = self.client.generate_content(
                full_prompt,
                generation_config=generation_config
            )
            
            content = response.text
            
            # Estimate tokens (Gemini doesn't provide exact counts in all cases)
            # Rough estimate: 1 token ≈ 4 characters
            estimated_tokens = (len(full_prompt) + len(content)) // 4
            
            # Gemini pricing (approximate, as of 2024)
            # Pro: $0.35/MTok input, $1.05/MTok output
            input_tokens = len(full_prompt) // 4
            output_tokens = len(content) // 4
            cost = (input_tokens * 0.35 / 1_000_000) + (output_tokens * 1.05 / 1_000_000)
            
            self.total_tokens += estimated_tokens
            self.total_cost += cost
            self.request_count += 1
            self.last_request_time = time.time()
            
            return CompletionResult(
                content=content,
                tokens_used=estimated_tokens,
                cost_usd=cost,
                model=self.model,
                finish_reason=(
                    response.candidates[0].finish_reason.name
                    if response.candidates
                    else "complete"
                )
            )
        
        try:
            result = self._retry_with_backoff(_make_request)
            self.logger.info(
                f"Completion successful: {result.tokens_used} tokens (est), "
                f"${result.cost_usd:.4f}"
            )
            return result
        except Exception as e:
            self.logger.error(f"Gemini API error: {e}", exc_info=True)
            raise


class OpenAIClient(LLMClient):
    """Client for OpenAI models (GPT-4, Codex, etc.)."""
    
    def __init__(self, model: str = "gpt-4"):
        super().__init__(model)
        self.api_key = os.getenv("OPENAI_API_KEY")
        if not self.api_key:
            raise ValueError("OPENAI_API_KEY environment variable not set")
        
        try:
            from openai import OpenAI
            self.client = OpenAI(api_key=self.api_key)
        except ImportError:
            raise ImportError("openai package not installed. Run: pip install openai")
    
    def complete(
        self,
        prompt: str,
        system_prompt: Optional[str] = None,
        max_tokens: int = 4096,
        temperature: float = 0.7
    ) -> CompletionResult:
        """Generate completion using OpenAI."""
        
        def _make_request():
            messages = []
            
            if system_prompt:
                messages.append({"role": "system", "content": system_prompt})
            
            messages.append({"role": "user", "content": prompt})
            
            response = self.client.chat.completions.create(
                model=self.model,
                messages=messages,
                max_tokens=max_tokens,
                temperature=temperature
            )
            
            content = response.choices[0].message.content
            
            # Get token usage
            total_tokens = response.usage.total_tokens
            prompt_tokens = response.usage.prompt_tokens
            completion_tokens = response.usage.completion_tokens
            
            # OpenAI pricing (approximate, varies by model)
            # GPT-4: $30/MTok input, $60/MTok output
            # GPT-3.5: $0.50/MTok input, $1.50/MTok output
            if "gpt-4" in self.model.lower():
                cost = (prompt_tokens * 30.0 / 1_000_000) + (completion_tokens * 60.0 / 1_000_000)
            else:
                cost = (prompt_tokens * 0.50 / 1_000_000) + (completion_tokens * 1.50 / 1_000_000)
            
            self.total_tokens += total_tokens
            self.total_cost += cost
            self.request_count += 1
            
            return CompletionResult(
                content=content,
                tokens_used=total_tokens,
                cost_usd=cost,
                model=self.model,
                finish_reason=response.choices[0].finish_reason
            )
        
        try:
            result = self._retry_with_backoff(_make_request)
            self.logger.info(
                f"Completion successful: {result.tokens_used} tokens, "
                f"${result.cost_usd:.4f}"
            )
            return result
        except Exception as e:
            self.logger.error(f"OpenAI API error: {e}", exc_info=True)
            raise


def create_llm_client(model: str) -> LLMClient:
    """
    Factory function to create the appropriate LLM client.
    
    Args:
        model: Model identifier (e.g., "claude", "gemini", "gpt-4")
        
    Returns:
        Appropriate LLMClient instance
    """
    model_lower = model.lower()
    
    if "claude" in model_lower:
        return ClaudeClient(model if model_lower != "claude" else "claude-3-5-sonnet-20241022")
    elif "gemini" in model_lower:
        return GeminiClient(model if model_lower != "gemini" else "gemini-1.5-pro")
    elif "gpt" in model_lower or "codex" in model_lower or "openai" in model_lower:
        return OpenAIClient(model if model_lower != "openai" else "gpt-4")
    else:
        raise ValueError(
            f"Unknown model: {model}. Supported: claude, gemini, gpt-4, codex"
        )
