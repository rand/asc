"""
Agent Adapter Package - AI coding agent framework for ASC.

This package provides the core functionality for headless AI agents that
execute development tasks using LLM providers.
"""

__version__ = "0.1.0"

from agent.llm_client import LLMClient, create_llm_client
from agent.phase_loop import HephaestusLoop
from agent.heartbeat import HeartbeatManager
from agent.ace import ACEPlaybook

__all__ = [
    "LLMClient",
    "create_llm_client",
    "HephaestusLoop",
    "HeartbeatManager",
    "ACEPlaybook",
]
