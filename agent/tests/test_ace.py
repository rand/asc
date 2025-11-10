"""Unit tests for ACE playbook system."""

import json
import tempfile
from pathlib import Path
from unittest.mock import Mock

from agent.ace import ACEPlaybook, Lesson


class TestACEPlaybook:
    """Test ACE playbook functionality."""
    
    def test_initialization(self, tmp_path, monkeypatch):
        """Test playbook initialization."""
        monkeypatch.setenv("HOME", str(tmp_path))
        
        playbook = ACEPlaybook("test-agent")
        
        assert playbook.agent_name == "test-agent"
        assert playbook.max_lessons == 100
        assert len(playbook.lessons) == 0
        assert playbook.playbook_dir.exists()
    
    def test_categorize_task(self, tmp_path, monkeypatch):
        """Test task categorization."""
        monkeypatch.setenv("HOME", str(tmp_path))
        playbook = ACEPlaybook("test-agent")
        
        # Create mock tasks
        test_task = Mock()
        test_task.phase = "testing"
        test_task.title = "Write unit tests"
        assert playbook._categorize_task(test_task) == "testing"
        
        impl_task = Mock()
        impl_task.phase = "implementation"
        impl_task.title = "Add feature"
        assert playbook._categorize_task(impl_task) == "implementation"
        
        bug_task = Mock()
        bug_task.phase = "general"
        bug_task.title = "Fix bug in parser"
        assert playbook._categorize_task(bug_task) == "bugfix"
    
    def test_add_lesson(self, tmp_path, monkeypatch):
        """Test adding lessons."""
        monkeypatch.setenv("HOME", str(tmp_path))
        playbook = ACEPlaybook("test-agent")
        
        lesson = Lesson(
            lesson_id="test123",
            context="Test context",
            action="Test action",
            outcome="success",
            learned="Test learning",
            task_type="testing",
            relevance_score=1.0,
            created_at="2024-11-09T10:00:00"
        )
        
        playbook._add_lesson(lesson)
        assert len(playbook.lessons) == 1
        assert playbook.lessons[0].lesson_id == "test123"
    
    def test_lessons_similar(self, tmp_path, monkeypatch):
        """Test lesson similarity detection."""
        monkeypatch.setenv("HOME", str(tmp_path))
        playbook = ACEPlaybook("test-agent")
        
        lesson1 = Lesson(
            lesson_id="1",
            context="implement user authentication system",
            action="action",
            outcome="success",
            learned="learned",
            task_type="implementation",
            relevance_score=1.0,
            created_at="2024-11-09T10:00:00"
        )
        
        lesson2 = Lesson(
            lesson_id="2",
            context="implement user authentication feature",
            action="action",
            outcome="success",
            learned="learned",
            task_type="implementation",
            relevance_score=1.0,
            created_at="2024-11-09T10:00:00"
        )
        
        assert playbook._lessons_similar(lesson1, lesson2)
    
    def test_get_relevant_lessons(self, tmp_path, monkeypatch):
        """Test retrieving relevant lessons."""
        monkeypatch.setenv("HOME", str(tmp_path))
        playbook = ACEPlaybook("test-agent")
        
        # Add some lessons
        for i in range(5):
            lesson = Lesson(
                lesson_id=f"test{i}",
                context=f"Test context {i}",
                action="Test action",
                outcome="success",
                learned="Test learning",
                task_type="implementation" if i < 3 else "testing",
                relevance_score=1.0 + i * 0.1,
                created_at="2024-11-09T10:00:00"
            )
            playbook.lessons.append(lesson)
        
        # Get relevant lessons for implementation
        relevant = playbook.get_relevant_lessons(
            "implementation",
            "Test context",
            max_lessons=3
        )
        
        assert len(relevant) <= 3
        assert all(isinstance(l, dict) for l in relevant)
    
    def test_save_and_load_playbook(self, tmp_path, monkeypatch):
        """Test saving and loading playbook."""
        monkeypatch.setenv("HOME", str(tmp_path))
        
        # Create and save playbook
        playbook1 = ACEPlaybook("test-agent")
        lesson = Lesson(
            lesson_id="test123",
            context="Test context",
            action="Test action",
            outcome="success",
            learned="Test learning",
            task_type="testing",
            relevance_score=1.0,
            created_at="2024-11-09T10:00:00"
        )
        playbook1.lessons.append(lesson)
        playbook1._save_playbook()
        
        # Load playbook
        playbook2 = ACEPlaybook("test-agent")
        assert len(playbook2.lessons) == 1
        assert playbook2.lessons[0].lesson_id == "test123"
    
    def test_curate_playbook(self, tmp_path, monkeypatch):
        """Test playbook curation."""
        monkeypatch.setenv("HOME", str(tmp_path))
        playbook = ACEPlaybook("test-agent", max_lessons=5)
        
        # Add more lessons than max
        for i in range(10):
            lesson = Lesson(
                lesson_id=f"test{i}",
                context=f"Test context {i}",
                action="Test action",
                outcome="success",
                learned="Test learning",
                task_type="testing",
                relevance_score=1.0 + i * 0.1,
                created_at="2024-11-09T10:00:00"
            )
            playbook.lessons.append(lesson)
        
        playbook._curate_playbook()
        
        # Should be pruned to max_lessons
        assert len(playbook.lessons) == 5
        
        # Should be sorted by relevance (highest first)
        scores = [l.relevance_score for l in playbook.lessons]
        assert scores == sorted(scores, reverse=True)
    
    def test_get_stats(self, tmp_path, monkeypatch):
        """Test playbook statistics."""
        monkeypatch.setenv("HOME", str(tmp_path))
        playbook = ACEPlaybook("test-agent")
        
        # Add lessons of different types
        for task_type in ["implementation", "testing", "implementation"]:
            lesson = Lesson(
                lesson_id=f"test-{task_type}",
                context="Test context",
                action="Test action",
                outcome="success",
                learned="Test learning",
                task_type=task_type,
                relevance_score=1.0,
                created_at="2024-11-09T10:00:00"
            )
            playbook.lessons.append(lesson)
        
        stats = playbook.get_stats()
        
        assert stats["total_lessons"] == 3
        assert stats["by_type"]["implementation"] == 2
        assert stats["by_type"]["testing"] == 1
        assert "avg_relevance" in stats
