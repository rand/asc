"""Unit tests for Hephaestus phase loop."""

import json
import logging
from pathlib import Path
from unittest.mock import Mock, patch, MagicMock

from agent.phase_loop import HephaestusLoop, Task, FileLease


class TestHephaestusLoop:
    """Test Hephaestus phase loop functionality."""
    
    def test_initialization(self, tmp_path):
        """Test phase loop initialization."""
        logger = logging.getLogger("test")
        llm_client = Mock()
        playbook = Mock()
        heartbeat = Mock()
        
        loop = HephaestusLoop(
            agent_name="test-agent",
            phases=["implementation", "testing"],
            llm_client=llm_client,
            playbook=playbook,
            beads_db_path=str(tmp_path),
            mcp_url="http://localhost:8765",
            heartbeat_manager=heartbeat,
            logger=logger
        )
        
        assert loop.agent_name == "test-agent"
        assert loop.phases == ["implementation", "testing"]
        assert loop.current_task is None
        assert len(loop.active_leases) == 0
    
    @patch('subprocess.run')
    def test_poll_for_task_success(self, mock_run, tmp_path):
        """Test successful task polling."""
        logger = logging.getLogger("test")
        llm_client = Mock()
        playbook = Mock()
        heartbeat = Mock()
        
        loop = HephaestusLoop(
            agent_name="test-agent",
            phases=["implementation"],
            llm_client=llm_client,
            playbook=playbook,
            beads_db_path=str(tmp_path),
            mcp_url="http://localhost:8765",
            heartbeat_manager=heartbeat,
            logger=logger
        )
        
        # Mock bd list output
        mock_result = Mock()
        mock_result.returncode = 0
        mock_result.stdout = json.dumps([
            {
                "id": "task-123",
                "title": "Test task",
                "status": "open",
                "phase": "implementation",
                "description": "Test description"
            }
        ])
        mock_run.return_value = mock_result
        
        task = loop._poll_for_task()
        
        assert task is not None
        assert task.id == "task-123"
        assert task.title == "Test task"
        assert task.phase == "implementation"
    
    @patch('subprocess.run')
    def test_poll_for_task_no_match(self, mock_run, tmp_path):
        """Test task polling with no matching phase."""
        logger = logging.getLogger("test")
        llm_client = Mock()
        playbook = Mock()
        heartbeat = Mock()
        
        loop = HephaestusLoop(
            agent_name="test-agent",
            phases=["testing"],
            llm_client=llm_client,
            playbook=playbook,
            beads_db_path=str(tmp_path),
            mcp_url="http://localhost:8765",
            heartbeat_manager=heartbeat,
            logger=logger
        )
        
        # Mock bd list output with different phase
        mock_result = Mock()
        mock_result.returncode = 0
        mock_result.stdout = json.dumps([
            {
                "id": "task-123",
                "title": "Test task",
                "status": "open",
                "phase": "implementation",
                "description": "Test description"
            }
        ])
        mock_run.return_value = mock_result
        
        task = loop._poll_for_task()
        
        assert task is None
    
    def test_identify_files_for_task(self, tmp_path):
        """Test file identification from task description."""
        logger = logging.getLogger("test")
        llm_client = Mock()
        playbook = Mock()
        heartbeat = Mock()
        
        loop = HephaestusLoop(
            agent_name="test-agent",
            phases=["implementation"],
            llm_client=llm_client,
            playbook=playbook,
            beads_db_path=str(tmp_path),
            mcp_url="http://localhost:8765",
            heartbeat_manager=heartbeat,
            logger=logger
        )
        
        task = Task(
            id="task-123",
            title="Test task",
            status="open",
            phase="implementation",
            description="Update `src/main.py` and file: tests/test_main.py"
        )
        
        files = loop._identify_files_for_task(task)
        
        assert "src/main.py" in files
        assert "tests/test_main.py" in files
    
    @patch('requests.post')
    def test_request_file_leases(self, mock_post, tmp_path):
        """Test file lease requests."""
        logger = logging.getLogger("test")
        llm_client = Mock()
        playbook = Mock()
        heartbeat = Mock()
        
        loop = HephaestusLoop(
            agent_name="test-agent",
            phases=["implementation"],
            llm_client=llm_client,
            playbook=playbook,
            beads_db_path=str(tmp_path),
            mcp_url="http://localhost:8765",
            heartbeat_manager=heartbeat,
            logger=logger
        )
        
        # Mock successful lease response
        mock_response = Mock()
        mock_response.status_code = 200
        mock_response.json.return_value = {"lease_id": "lease-123"}
        mock_post.return_value = mock_response
        
        leases = loop._request_file_leases(["src/main.py"])
        
        assert len(leases) == 1
        assert leases[0].lease_id == "lease-123"
        assert leases[0].file_path == "src/main.py"
    
    def test_build_context(self, tmp_path):
        """Test context building."""
        logger = logging.getLogger("test")
        llm_client = Mock()
        playbook = Mock()
        heartbeat = Mock()
        
        loop = HephaestusLoop(
            agent_name="test-agent",
            phases=["implementation"],
            llm_client=llm_client,
            playbook=playbook,
            beads_db_path=str(tmp_path),
            mcp_url="http://localhost:8765",
            heartbeat_manager=heartbeat,
            logger=logger
        )
        
        # Create a test file
        test_file = tmp_path / "test.py"
        test_file.write_text("print('hello')")
        
        task = Task(
            id="task-123",
            title="Test task",
            status="open",
            phase="implementation",
            description="Test description"
        )
        
        lease = FileLease(
            lease_id="lease-123",
            file_path="test.py",
            agent_name="test-agent"
        )
        
        context = loop._build_context(task, [lease])
        
        assert context["task_id"] == "task-123"
        assert context["task_title"] == "Test task"
        assert "test.py" in context["files"]
        assert context["files"]["test.py"] == "print('hello')"
    
    def test_generate_prompt(self, tmp_path):
        """Test prompt generation."""
        logger = logging.getLogger("test")
        llm_client = Mock()
        playbook = Mock()
        heartbeat = Mock()
        
        loop = HephaestusLoop(
            agent_name="test-agent",
            phases=["implementation"],
            llm_client=llm_client,
            playbook=playbook,
            beads_db_path=str(tmp_path),
            mcp_url="http://localhost:8765",
            heartbeat_manager=heartbeat,
            logger=logger
        )
        
        task = Task(
            id="task-123",
            title="Test task",
            status="open",
            phase="implementation",
            description="Test description"
        )
        
        context = {
            "task_id": "task-123",
            "task_title": "Test task",
            "task_description": "Test description",
            "phase": "implementation",
            "files": {"test.py": "print('hello')"}
        }
        
        lessons = [
            {
                "context": "Similar task",
                "action": "Did something",
                "outcome": "success",
                "learned": "Learned something"
            }
        ]
        
        prompt = loop._generate_prompt(task, context, lessons)
        
        assert "Test task" in prompt
        assert "Test description" in prompt
        assert "test.py" in prompt
        assert "Similar task" in prompt
    
    @patch('subprocess.run')
    def test_update_task_status(self, mock_run, tmp_path):
        """Test task status updates."""
        logger = logging.getLogger("test")
        llm_client = Mock()
        playbook = Mock()
        heartbeat = Mock()
        
        loop = HephaestusLoop(
            agent_name="test-agent",
            phases=["implementation"],
            llm_client=llm_client,
            playbook=playbook,
            beads_db_path=str(tmp_path),
            mcp_url="http://localhost:8765",
            heartbeat_manager=heartbeat,
            logger=logger
        )
        
        mock_result = Mock()
        mock_result.returncode = 0
        mock_run.return_value = mock_result
        
        loop._update_task_status("task-123", "in_progress")
        
        assert mock_run.called
        call_args = mock_run.call_args[0][0]
        assert "bd" in call_args
        assert "update" in call_args
        assert "task-123" in call_args
        assert "in_progress" in call_args
    
    @patch('requests.post')
    def test_release_all_leases(self, mock_post, tmp_path):
        """Test releasing all leases."""
        logger = logging.getLogger("test")
        llm_client = Mock()
        playbook = Mock()
        heartbeat = Mock()
        
        loop = HephaestusLoop(
            agent_name="test-agent",
            phases=["implementation"],
            llm_client=llm_client,
            playbook=playbook,
            beads_db_path=str(tmp_path),
            mcp_url="http://localhost:8765",
            heartbeat_manager=heartbeat,
            logger=logger
        )
        
        # Add some leases
        loop.active_leases = [
            FileLease("lease-1", "file1.py", "test-agent"),
            FileLease("lease-2", "file2.py", "test-agent")
        ]
        
        mock_response = Mock()
        mock_response.status_code = 200
        mock_post.return_value = mock_response
        
        loop._release_all_leases()
        
        assert len(loop.active_leases) == 0
        assert mock_post.call_count == 2
