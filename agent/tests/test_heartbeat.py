"""Unit tests for heartbeat system."""

import time
import logging
from unittest.mock import Mock, patch, MagicMock

from agent.heartbeat import HeartbeatManager


class TestHeartbeatManager:
    """Test heartbeat manager functionality."""
    
    def test_initialization(self):
        """Test heartbeat manager initialization."""
        logger = logging.getLogger("test")
        manager = HeartbeatManager(
            agent_name="test-agent",
            mcp_url="http://localhost:8765",
            logger=logger,
            interval=30
        )
        
        assert manager.agent_name == "test-agent"
        assert manager.mcp_url == "http://localhost:8765"
        assert manager.interval == 30
        assert manager.status == "idle"
        assert manager.current_task is None
        assert not manager.running
    
    def test_update_status(self):
        """Test status updates."""
        logger = logging.getLogger("test")
        manager = HeartbeatManager(
            agent_name="test-agent",
            mcp_url="http://localhost:8765",
            logger=logger
        )
        
        with patch.object(manager, '_send_heartbeat') as mock_send:
            manager.update_status("working", current_task="task-123")
            
            assert manager.status == "working"
            assert manager.current_task == "task-123"
            mock_send.assert_called_once()
    
    def test_update_status_no_change(self):
        """Test status update with no change."""
        logger = logging.getLogger("test")
        manager = HeartbeatManager(
            agent_name="test-agent",
            mcp_url="http://localhost:8765",
            logger=logger
        )
        manager.status = "idle"
        
        with patch.object(manager, '_send_heartbeat') as mock_send:
            manager.update_status("idle")
            
            # Should not send heartbeat if status unchanged
            mock_send.assert_not_called()
    
    @patch('requests.post')
    def test_send_heartbeat_success(self, mock_post):
        """Test successful heartbeat send."""
        mock_response = Mock()
        mock_response.status_code = 200
        mock_post.return_value = mock_response
        
        logger = logging.getLogger("test")
        manager = HeartbeatManager(
            agent_name="test-agent",
            mcp_url="http://localhost:8765",
            logger=logger
        )
        
        manager._send_heartbeat()
        
        assert mock_post.called
        assert manager.last_heartbeat is not None
        assert manager.backoff_time == 1
    
    @patch('requests.post')
    def test_send_heartbeat_failure(self, mock_post):
        """Test heartbeat send failure."""
        mock_post.side_effect = Exception("Connection error")
        
        logger = logging.getLogger("test")
        manager = HeartbeatManager(
            agent_name="test-agent",
            mcp_url="http://localhost:8765",
            logger=logger
        )
        
        with patch.object(manager, '_handle_connection_failure') as mock_handle:
            manager._send_heartbeat()
            mock_handle.assert_called_once()
    
    def test_handle_connection_failure(self):
        """Test connection failure handling."""
        logger = logging.getLogger("test")
        manager = HeartbeatManager(
            agent_name="test-agent",
            mcp_url="http://localhost:8765",
            logger=logger
        )
        
        initial_backoff = manager.backoff_time
        
        with patch('time.sleep'):
            manager._handle_connection_failure()
        
        # Backoff should increase
        assert manager.backoff_time > initial_backoff
    
    def test_is_healthy(self):
        """Test health check."""
        logger = logging.getLogger("test")
        manager = HeartbeatManager(
            agent_name="test-agent",
            mcp_url="http://localhost:8765",
            logger=logger
        )
        
        # No heartbeat yet
        assert not manager.is_healthy()
        
        # Set recent heartbeat
        from datetime import datetime
        manager.last_heartbeat = datetime.now()
        assert manager.is_healthy()
    
    def test_get_stats(self):
        """Test statistics retrieval."""
        logger = logging.getLogger("test")
        manager = HeartbeatManager(
            agent_name="test-agent",
            mcp_url="http://localhost:8765",
            logger=logger
        )
        manager.status = "working"
        manager.current_task = "task-123"
        
        stats = manager.get_stats()
        
        assert stats["agent_name"] == "test-agent"
        assert stats["status"] == "working"
        assert stats["current_task"] == "task-123"
        assert "is_healthy" in stats
        assert "backoff_time" in stats
    
    @patch('requests.post')
    def test_start_and_stop(self, mock_post):
        """Test starting and stopping heartbeat thread."""
        mock_response = Mock()
        mock_response.status_code = 200
        mock_post.return_value = mock_response
        
        logger = logging.getLogger("test")
        manager = HeartbeatManager(
            agent_name="test-agent",
            mcp_url="http://localhost:8765",
            logger=logger,
            interval=1  # Short interval for testing
        )
        
        # Start
        manager.start()
        assert manager.running
        assert manager.thread is not None
        
        # Let it run briefly
        time.sleep(0.5)
        
        # Stop
        manager.stop()
        assert not manager.running
