#!/usr/bin/env python3
"""
Agent Adapter - Entry point for AI coding agents in the ASC system.

This module serves as the main entry point for headless AI agents that execute
development tasks using LLM providers (Claude, Gemini, OpenAI).
"""

import os
import sys
import signal
import logging
import time
from pathlib import Path
from typing import Optional

from agent.llm_client import create_llm_client
from agent.phase_loop import HephaestusLoop
from agent.heartbeat import HeartbeatManager
from agent.ace import ACEPlaybook


class AgentAdapter:
    """Main agent adapter that orchestrates the agent lifecycle."""
    
    def __init__(self):
        self.agent_name = os.getenv("AGENT_NAME", "unknown-agent")
        self.agent_model = os.getenv("AGENT_MODEL", "claude")
        self.agent_phases = os.getenv("AGENT_PHASES", "").split(",")
        self.mcp_mail_url = os.getenv("MCP_MAIL_URL", "http://localhost:8765")
        self.beads_db_path = os.getenv("BEADS_DB_PATH", "./project-repo")
        
        self.running = True
        self.logger = self._setup_logging()
        self.llm_client = None
        self.phase_loop = None
        self.heartbeat_manager = None
        self.playbook = None
        
    def _setup_logging(self) -> logging.Logger:
        """Initialize logging to ~/.asc/logs/{agent_name}.log"""
        log_dir = Path.home() / ".asc" / "logs"
        log_dir.mkdir(parents=True, exist_ok=True)
        
        log_file = log_dir / f"{self.agent_name}.log"
        
        # Configure logging
        logging.basicConfig(
            level=logging.INFO,
            format='%(asctime)s [%(levelname)s] %(name)s: %(message)s',
            handlers=[
                logging.FileHandler(log_file),
                logging.StreamHandler(sys.stdout)
            ]
        )
        
        logger = logging.getLogger(f"agent.{self.agent_name}")
        logger.info(f"Agent {self.agent_name} starting up")
        logger.info(f"Model: {self.agent_model}")
        logger.info(f"Phases: {', '.join(self.agent_phases)}")
        
        return logger
    
    def _setup_signal_handlers(self):
        """Set up signal handlers for graceful shutdown."""
        def signal_handler(signum, frame):
            sig_name = signal.Signals(signum).name
            self.logger.info(f"Received {sig_name}, initiating graceful shutdown")
            self.running = False
        
        signal.signal(signal.SIGTERM, signal_handler)
        signal.signal(signal.SIGINT, signal_handler)
        
        self.logger.info("Signal handlers registered for SIGTERM and SIGINT")
    
    def _validate_environment(self) -> bool:
        """Validate required environment variables are present."""
        required_vars = {
            "AGENT_NAME": self.agent_name,
            "AGENT_MODEL": self.agent_model,
            "AGENT_PHASES": os.getenv("AGENT_PHASES"),
            "MCP_MAIL_URL": self.mcp_mail_url,
            "BEADS_DB_PATH": self.beads_db_path,
        }
        
        missing = []
        for var, value in required_vars.items():
            if not value or value == "unknown-agent":
                missing.append(var)
        
        if missing:
            self.logger.error(f"Missing required environment variables: {', '.join(missing)}")
            return False
        
        # Check for API keys based on model
        api_key_map = {
            "claude": "CLAUDE_API_KEY",
            "gemini": "GOOGLE_API_KEY",
            "openai": "OPENAI_API_KEY",
            "gpt-4": "OPENAI_API_KEY",
            "codex": "OPENAI_API_KEY",
        }
        
        required_key = api_key_map.get(self.agent_model.lower())
        if required_key and not os.getenv(required_key):
            self.logger.error(f"Missing API key: {required_key} for model {self.agent_model}")
            return False
        
        self.logger.info("Environment validation passed")
        return True
    
    def initialize(self) -> bool:
        """Initialize all agent components."""
        try:
            # Validate environment
            if not self._validate_environment():
                return False
            
            # Initialize LLM client
            self.logger.info(f"Initializing LLM client for model: {self.agent_model}")
            self.llm_client = create_llm_client(self.agent_model)
            
            # Initialize ACE playbook
            self.logger.info("Initializing ACE playbook")
            self.playbook = ACEPlaybook(self.agent_name)
            
            # Initialize heartbeat manager
            self.logger.info("Initializing heartbeat manager")
            self.heartbeat_manager = HeartbeatManager(
                agent_name=self.agent_name,
                mcp_url=self.mcp_mail_url,
                logger=self.logger
            )
            
            # Initialize phase loop
            self.logger.info("Initializing Hephaestus phase loop")
            self.phase_loop = HephaestusLoop(
                agent_name=self.agent_name,
                phases=self.agent_phases,
                llm_client=self.llm_client,
                playbook=self.playbook,
                beads_db_path=self.beads_db_path,
                mcp_url=self.mcp_mail_url,
                heartbeat_manager=self.heartbeat_manager,
                logger=self.logger
            )
            
            self.logger.info("Agent initialization complete")
            return True
            
        except Exception as e:
            self.logger.error(f"Failed to initialize agent: {e}", exc_info=True)
            return False
    
    def run(self):
        """Main event loop."""
        self.logger.info("Entering main event loop")
        
        # Start heartbeat
        self.heartbeat_manager.start()
        
        try:
            while self.running:
                try:
                    # Execute one iteration of the phase loop
                    self.phase_loop.iterate()
                    
                    # Brief sleep to prevent tight loop
                    time.sleep(1)
                    
                except KeyboardInterrupt:
                    self.logger.info("Keyboard interrupt received")
                    break
                except Exception as e:
                    self.logger.error(f"Error in main loop: {e}", exc_info=True)
                    self.heartbeat_manager.update_status("error", error=str(e))
                    time.sleep(5)  # Back off on errors
                    
        finally:
            self.shutdown()
    
    def shutdown(self):
        """Graceful shutdown of all components."""
        self.logger.info("Shutting down agent")
        
        if self.heartbeat_manager:
            self.heartbeat_manager.stop()
        
        if self.phase_loop:
            self.phase_loop.cleanup()
        
        self.logger.info("Agent shutdown complete")


def main():
    """Main entry point."""
    adapter = AgentAdapter()
    
    # Set up signal handlers
    adapter._setup_signal_handlers()
    
    # Initialize
    if not adapter.initialize():
        adapter.logger.error("Agent initialization failed, exiting")
        sys.exit(1)
    
    # Run
    try:
        adapter.run()
    except Exception as e:
        adapter.logger.error(f"Fatal error: {e}", exc_info=True)
        sys.exit(1)
    
    sys.exit(0)


if __name__ == "__main__":
    main()
