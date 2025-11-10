"""
Agent Heartbeat System - Periodic status reporting to MCP.

Sends heartbeat messages to mcp_agent_mail to report agent status,
enabling the TUI to display real-time agent state.
"""

import time
import logging
import requests
import threading
from typing import Optional
from datetime import datetime


class HeartbeatManager:
    """Manages periodic heartbeat messages to MCP."""
    
    def __init__(
        self,
        agent_name: str,
        mcp_url: str,
        logger: logging.Logger,
        interval: int = 30
    ):
        self.agent_name = agent_name
        self.mcp_url = mcp_url.rstrip("/")
        self.logger = logger
        self.interval = interval
        
        self.status = "idle"
        self.current_task: Optional[str] = None
        self.error: Optional[str] = None
        
        self.running = False
        self.thread: Optional[threading.Thread] = None
        self.last_heartbeat: Optional[datetime] = None
        
        # Exponential backoff for connection failures
        self.backoff_time = 1
        self.max_backoff = 300  # 5 minutes
        
        self.logger.info(
            f"Heartbeat manager initialized (interval: {interval}s)"
        )
    
    def start(self):
        """Start the heartbeat thread."""
        if self.running:
            self.logger.warning("Heartbeat already running")
            return
        
        self.running = True
        self.thread = threading.Thread(target=self._heartbeat_loop, daemon=True)
        self.thread.start()
        
        self.logger.info("Heartbeat thread started")
    
    def stop(self):
        """Stop the heartbeat thread."""
        if not self.running:
            return
        
        self.running = False
        
        if self.thread:
            self.thread.join(timeout=5)
        
        # Send final offline status
        self._send_heartbeat(status="offline")
        
        self.logger.info("Heartbeat thread stopped")
    
    def update_status(
        self,
        status: str,
        current_task: Optional[str] = None,
        error: Optional[str] = None
    ):
        """
        Update agent status and send immediate heartbeat.
        
        Args:
            status: Agent status (idle, working, error, offline)
            current_task: Current task ID if working
            error: Error message if status is error
        """
        # Track state transitions
        old_status = self.status
        self.status = status
        self.current_task = current_task
        self.error = error
        
        # Send immediate heartbeat on state change
        if old_status != status:
            self.logger.info(f"Status changed: {old_status} -> {status}")
            self._send_heartbeat()
    
    def _heartbeat_loop(self):
        """Main heartbeat loop (runs in separate thread)."""
        while self.running:
            try:
                self._send_heartbeat()
                
                # Sleep in small increments to allow quick shutdown
                for _ in range(self.interval):
                    if not self.running:
                        break
                    time.sleep(1)
                    
            except Exception as e:
                self.logger.error(f"Error in heartbeat loop: {e}", exc_info=True)
                time.sleep(5)
    
    def _send_heartbeat(self, status: Optional[str] = None):
        """Send a heartbeat message to MCP."""
        try:
            payload = {
                "agent_name": self.agent_name,
                "status": status or self.status,
                "timestamp": datetime.now().isoformat(),
            }
            
            if self.current_task:
                payload["current_task"] = self.current_task
            
            if self.error:
                payload["error"] = self.error
            
            response = requests.post(
                f"{self.mcp_url}/heartbeat",
                json=payload,
                timeout=5
            )
            
            if response.status_code == 200:
                self.last_heartbeat = datetime.now()
                self.backoff_time = 1  # Reset backoff on success
                
                self.logger.debug(
                    f"Heartbeat sent: {self.status}"
                    + (f" (task: {self.current_task})" if self.current_task else "")
                )
            else:
                self.logger.warning(
                    f"Heartbeat failed with status {response.status_code}: "
                    f"{response.text}"
                )
                self._handle_connection_failure()
                
        except requests.exceptions.Timeout:
            self.logger.warning("Heartbeat request timed out")
            self._handle_connection_failure()
            
        except requests.exceptions.ConnectionError as e:
            self.logger.warning(f"MCP connection error: {e}")
            self._handle_connection_failure()
            
        except Exception as e:
            self.logger.error(f"Error sending heartbeat: {e}", exc_info=True)
            self._handle_connection_failure()
    
    def _handle_connection_failure(self):
        """Handle MCP connection failures with exponential backoff."""
        # Increase backoff time
        self.backoff_time = min(self.backoff_time * 2, self.max_backoff)
        
        self.logger.info(
            f"MCP temporarily unavailable, backing off for {self.backoff_time}s"
        )
        
        # Sleep with backoff (but allow quick shutdown)
        for _ in range(self.backoff_time):
            if not self.running:
                break
            time.sleep(1)
    
    def is_healthy(self) -> bool:
        """Check if heartbeat is healthy."""
        if not self.last_heartbeat:
            return False
        
        # Consider unhealthy if no successful heartbeat in 5 minutes
        elapsed = (datetime.now() - self.last_heartbeat).total_seconds()
        return elapsed < 300
    
    def get_stats(self) -> dict:
        """Get heartbeat statistics."""
        return {
            "agent_name": self.agent_name,
            "status": self.status,
            "current_task": self.current_task,
            "last_heartbeat": (
                self.last_heartbeat.isoformat() if self.last_heartbeat else None
            ),
            "is_healthy": self.is_healthy(),
            "backoff_time": self.backoff_time
        }
