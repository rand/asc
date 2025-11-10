"""Unit tests for LLM client abstraction."""

import os
import pytest
from unittest.mock import Mock, patch, MagicMock

from agent.llm_client import (
    LLMClient,
    ClaudeClient,
    GeminiClient,
    OpenAIClient,
    create_llm_client,
    CompletionResult
)


class TestLLMClientBase:
    """Test base LLM client functionality."""
    
    def test_retry_with_backoff_success(self):
        """Test retry logic succeeds on first attempt."""
        
        class TestClient(LLMClient):
            def complete(self, prompt, system_prompt=None, max_tokens=4096, temperature=0.7):
                pass
        
        client = TestClient("test-model")
        
        mock_func = Mock(return_value="success")
        result = client._retry_with_backoff(mock_func, max_retries=3)
        
        assert result == "success"
        assert mock_func.call_count == 1
    
    def test_retry_with_backoff_eventual_success(self):
        """Test retry logic succeeds after failures."""
        
        class TestClient(LLMClient):
            def complete(self, prompt, system_prompt=None, max_tokens=4096, temperature=0.7):
                pass
        
        client = TestClient("test-model")
        
        mock_func = Mock(side_effect=[Exception("fail"), Exception("fail"), "success"])
        result = client._retry_with_backoff(mock_func, max_retries=3)
        
        assert result == "success"
        assert mock_func.call_count == 3
    
    def test_get_stats(self):
        """Test statistics tracking."""
        
        class TestClient(LLMClient):
            def complete(self, prompt, system_prompt=None, max_tokens=4096, temperature=0.7):
                pass
        
        client = TestClient("test-model")
        client.total_tokens = 1000
        client.total_cost = 0.05
        client.request_count = 5
        
        stats = client.get_stats()
        
        assert stats["model"] == "test-model"
        assert stats["total_tokens"] == 1000
        assert stats["total_cost_usd"] == 0.05
        assert stats["request_count"] == 5
        assert stats["avg_tokens_per_request"] == 200


class TestCreateLLMClient:
    """Test LLM client factory function."""
    
    @patch.dict(os.environ, {"CLAUDE_API_KEY": "test-key"})
    @patch("agent.llm_client.Anthropic")
    def test_create_claude_client(self, mock_anthropic):
        """Test creating Claude client."""
        client = create_llm_client("claude")
        assert isinstance(client, ClaudeClient)
    
    @patch.dict(os.environ, {"GOOGLE_API_KEY": "test-key"})
    @patch("agent.llm_client.genai")
    def test_create_gemini_client(self, mock_genai):
        """Test creating Gemini client."""
        client = create_llm_client("gemini")
        assert isinstance(client, GeminiClient)
    
    @patch.dict(os.environ, {"OPENAI_API_KEY": "test-key"})
    @patch("agent.llm_client.OpenAI")
    def test_create_openai_client(self, mock_openai):
        """Test creating OpenAI client."""
        client = create_llm_client("gpt-4")
        assert isinstance(client, OpenAIClient)
    
    def test_create_unknown_model(self):
        """Test error on unknown model."""
        with pytest.raises(ValueError, match="Unknown model"):
            create_llm_client("unknown-model")


class TestClaudeClient:
    """Test Claude client implementation."""
    
    @patch.dict(os.environ, {"CLAUDE_API_KEY": "test-key"})
    @patch("agent.llm_client.Anthropic")
    def test_initialization(self, mock_anthropic):
        """Test Claude client initialization."""
        client = ClaudeClient()
        assert client.model == "claude-3-5-sonnet-20241022"
        assert client.api_key == "test-key"
    
    def test_missing_api_key(self):
        """Test error when API key is missing."""
        with patch.dict(os.environ, {}, clear=True):
            with pytest.raises(ValueError, match="CLAUDE_API_KEY"):
                ClaudeClient()


class TestGeminiClient:
    """Test Gemini client implementation."""
    
    @patch.dict(os.environ, {"GOOGLE_API_KEY": "test-key"})
    @patch("agent.llm_client.genai")
    def test_initialization(self, mock_genai):
        """Test Gemini client initialization."""
        client = GeminiClient()
        assert client.model == "gemini-1.5-pro"
        assert client.api_key == "test-key"
    
    def test_missing_api_key(self):
        """Test error when API key is missing."""
        with patch.dict(os.environ, {}, clear=True):
            with pytest.raises(ValueError, match="GOOGLE_API_KEY"):
                GeminiClient()


class TestOpenAIClient:
    """Test OpenAI client implementation."""
    
    @patch.dict(os.environ, {"OPENAI_API_KEY": "test-key"})
    @patch("agent.llm_client.OpenAI")
    def test_initialization(self, mock_openai):
        """Test OpenAI client initialization."""
        client = OpenAIClient()
        assert client.model == "gpt-4"
        assert client.api_key == "test-key"
    
    def test_missing_api_key(self):
        """Test error when API key is missing."""
        with patch.dict(os.environ, {}, clear=True):
            with pytest.raises(ValueError, match="OPENAI_API_KEY"):
                OpenAIClient()
