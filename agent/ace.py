"""
ACE (Agentic Context Engineering) - Playbook-based learning system.

Agents reflect on completed tasks and build a playbook of lessons learned
to improve future performance.
"""

import json
import logging
import hashlib
from pathlib import Path
from typing import List, Dict, Any, Optional
from dataclasses import dataclass, asdict
from datetime import datetime


@dataclass
class Lesson:
    """Represents a learned lesson in the playbook."""
    lesson_id: str
    context: str
    action: str
    outcome: str
    learned: str
    task_type: str
    relevance_score: float
    created_at: str
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary."""
        return asdict(self)
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> 'Lesson':
        """Create from dictionary."""
        return cls(**data)


class ACEPlaybook:
    """Manages the agent's playbook of learned lessons."""
    
    def __init__(self, agent_name: str, max_lessons: int = 100):
        self.agent_name = agent_name
        self.max_lessons = max_lessons
        self.logger = logging.getLogger(f"ace.{agent_name}")
        
        # Set up playbook directory
        self.playbook_dir = Path.home() / ".asc" / "playbooks" / agent_name
        self.playbook_dir.mkdir(parents=True, exist_ok=True)
        
        self.playbook_file = self.playbook_dir / "playbook.json"
        self.lessons: List[Lesson] = []
        
        # Load existing playbook
        self._load_playbook()
        
        self.logger.info(
            f"ACE playbook initialized with {len(self.lessons)} lessons"
        )
    
    def _load_playbook(self):
        """Load playbook from disk."""
        if not self.playbook_file.exists():
            self.logger.info("No existing playbook found, starting fresh")
            return
        
        try:
            with open(self.playbook_file, 'r', encoding='utf-8') as f:
                data = json.load(f)
                self.lessons = [Lesson.from_dict(l) for l in data.get("lessons", [])]
            
            self.logger.info(f"Loaded {len(self.lessons)} lessons from playbook")
            
        except Exception as e:
            self.logger.error(f"Error loading playbook: {e}", exc_info=True)
            self.lessons = []
    
    def _save_playbook(self):
        """Save playbook to disk."""
        try:
            data = {
                "agent_name": self.agent_name,
                "updated_at": datetime.now().isoformat(),
                "lessons": [l.to_dict() for l in self.lessons]
            }
            
            with open(self.playbook_file, 'w', encoding='utf-8') as f:
                json.dump(data, f, indent=2)
            
            self.logger.info(f"Saved {len(self.lessons)} lessons to playbook")
            
        except Exception as e:
            self.logger.error(f"Error saving playbook: {e}", exc_info=True)
    
    def reflect_on_task(self, task, llm_response: str, outcome: str):
        """
        Reflect on a completed task and extract lessons.
        
        Args:
            task: The completed task
            llm_response: The LLM's response/action plan
            outcome: The outcome (e.g., "success", "error: ...")
        """
        try:
            self.logger.info(f"Reflecting on task {task.id}")
            
            # Generate reflection prompt
            reflection_prompt = self._generate_reflection_prompt(
                task, llm_response, outcome
            )
            
            # For now, extract lessons from the task execution
            # In a full implementation, we'd call the LLM again for reflection
            lesson = self._extract_lesson_from_task(task, llm_response, outcome)
            
            if lesson:
                self._add_lesson(lesson)
                self._curate_playbook()
                self._save_playbook()
            
        except Exception as e:
            self.logger.error(f"Error reflecting on task: {e}", exc_info=True)
    
    def _generate_reflection_prompt(
        self,
        task,
        llm_response: str,
        outcome: str
    ) -> str:
        """Generate a reflection prompt for the LLM."""
        return f"""# Task Reflection

## Task Details
- ID: {task.id}
- Title: {task.title}
- Phase: {task.phase}
- Description: {task.description}

## Your Response
{llm_response[:1000]}...

## Outcome
{outcome}

## Reflection Questions
1. What was the key challenge in this task?
2. What approach did you take?
3. What worked well?
4. What could be improved?
5. What general lesson can be applied to similar tasks?

Provide your reflection in JSON format:
{{
  "context": "Brief description of the task context",
  "action": "What you did",
  "outcome": "What happened",
  "learned": "Key lesson for future tasks"
}}
"""
    
    def _extract_lesson_from_task(
        self,
        task,
        llm_response: str,
        outcome: str
    ) -> Optional[Lesson]:
        """Extract a lesson from task execution (simplified version)."""
        # Generate a unique lesson ID
        lesson_id = hashlib.md5(
            f"{task.id}{task.phase}{datetime.now().isoformat()}".encode()
        ).hexdigest()[:12]
        
        # Determine task type from phase and title
        task_type = self._categorize_task(task)
        
        # Create lesson
        lesson = Lesson(
            lesson_id=lesson_id,
            context=f"Task in {task.phase} phase: {task.title}",
            action=f"Executed task with {len(llm_response)} char response",
            outcome=outcome,
            learned=self._extract_key_learning(task, outcome),
            task_type=task_type,
            relevance_score=1.0,  # Initial score
            created_at=datetime.now().isoformat()
        )
        
        return lesson
    
    def _categorize_task(self, task) -> str:
        """Categorize task type based on phase and title."""
        phase = task.phase.lower()
        title = task.title.lower()
        
        # Simple categorization
        if "test" in phase or "test" in title:
            return "testing"
        elif "implement" in phase or "code" in title or "add" in title:
            return "implementation"
        elif "plan" in phase or "design" in title:
            return "planning"
        elif "refactor" in phase or "refactor" in title:
            return "refactoring"
        elif "bug" in title or "fix" in title:
            return "bugfix"
        else:
            return "general"
    
    def _extract_key_learning(self, task, outcome: str) -> str:
        """Extract key learning from task outcome."""
        if outcome == "success":
            return f"Successfully completed {task.phase} task: {task.title}"
        elif outcome.startswith("error"):
            error_msg = outcome.replace("error: ", "")
            return f"Encountered error in {task.phase}: {error_msg}. Need better error handling."
        else:
            return f"Completed task with outcome: {outcome}"
    
    def _add_lesson(self, lesson: Lesson):
        """Add a lesson to the playbook."""
        # Check for duplicates
        for existing in self.lessons:
            if self._lessons_similar(existing, lesson):
                self.logger.info(
                    f"Lesson similar to {existing.lesson_id}, merging"
                )
                self._merge_lessons(existing, lesson)
                return
        
        # Add new lesson
        self.lessons.append(lesson)
        self.logger.info(f"Added new lesson {lesson.lesson_id}")
    
    def _lessons_similar(self, lesson1: Lesson, lesson2: Lesson) -> bool:
        """Check if two lessons are similar."""
        # Simple similarity check based on task type and context
        if lesson1.task_type != lesson2.task_type:
            return False
        
        # Check context similarity (simple word overlap)
        words1 = set(lesson1.context.lower().split())
        words2 = set(lesson2.context.lower().split())
        
        if len(words1) == 0 or len(words2) == 0:
            return False
        
        overlap = len(words1 & words2) / len(words1 | words2)
        return overlap > 0.6
    
    def _merge_lessons(self, existing: Lesson, new: Lesson):
        """Merge a new lesson into an existing one."""
        # Update relevance score (increase for repeated patterns)
        existing.relevance_score = min(existing.relevance_score + 0.1, 2.0)
        
        # Update learned field to include new insights
        if new.learned not in existing.learned:
            existing.learned += f" | {new.learned}"
    
    def _curate_playbook(self):
        """Curate playbook by deduplicating and pruning."""
        # Sort by relevance score (descending)
        self.lessons.sort(key=lambda l: l.relevance_score, reverse=True)
        
        # Prune to max lessons
        if len(self.lessons) > self.max_lessons:
            removed = len(self.lessons) - self.max_lessons
            self.lessons = self.lessons[:self.max_lessons]
            self.logger.info(f"Pruned {removed} lessons from playbook")
        
        # Decay relevance scores over time
        for lesson in self.lessons:
            lesson.relevance_score *= 0.99
    
    def get_relevant_lessons(
        self,
        phase: str,
        task_description: str,
        max_lessons: int = 5
    ) -> List[Dict[str, Any]]:
        """
        Get relevant lessons for a task.
        
        Args:
            phase: Task phase
            task_description: Task description
            max_lessons: Maximum number of lessons to return
            
        Returns:
            List of relevant lessons as dictionaries
        """
        if not self.lessons:
            return []
        
        # Score lessons by relevance
        scored_lessons = []
        
        for lesson in self.lessons:
            score = self._calculate_relevance(lesson, phase, task_description)
            scored_lessons.append((score, lesson))
        
        # Sort by score (descending)
        scored_lessons.sort(key=lambda x: x[0], reverse=True)
        
        # Return top lessons
        relevant = [l.to_dict() for _, l in scored_lessons[:max_lessons]]
        
        self.logger.info(
            f"Retrieved {len(relevant)} relevant lessons for {phase} phase"
        )
        
        return relevant
    
    def _calculate_relevance(
        self,
        lesson: Lesson,
        phase: str,
        task_description: str
    ) -> float:
        """Calculate relevance score for a lesson."""
        score = lesson.relevance_score
        
        # Boost for matching task type
        task_type = self._categorize_task_from_description(phase, task_description)
        if lesson.task_type == task_type:
            score *= 1.5
        
        # Boost for keyword matches in context
        desc_words = set(task_description.lower().split())
        context_words = set(lesson.context.lower().split())
        overlap = len(desc_words & context_words)
        score += overlap * 0.1
        
        return score
    
    def _categorize_task_from_description(
        self,
        phase: str,
        description: str
    ) -> str:
        """Categorize task from phase and description."""
        phase_lower = phase.lower()
        desc_lower = description.lower()
        
        if "test" in phase_lower or "test" in desc_lower:
            return "testing"
        elif "implement" in phase_lower or "code" in desc_lower:
            return "implementation"
        elif "plan" in phase_lower or "design" in desc_lower:
            return "planning"
        elif "refactor" in phase_lower or "refactor" in desc_lower:
            return "refactoring"
        elif "bug" in desc_lower or "fix" in desc_lower:
            return "bugfix"
        else:
            return "general"
    
    def get_stats(self) -> Dict[str, Any]:
        """Get playbook statistics."""
        if not self.lessons:
            return {
                "total_lessons": 0,
                "by_type": {},
                "avg_relevance": 0.0
            }
        
        by_type = {}
        for lesson in self.lessons:
            by_type[lesson.task_type] = by_type.get(lesson.task_type, 0) + 1
        
        avg_relevance = sum(l.relevance_score for l in self.lessons) / len(self.lessons)
        
        return {
            "total_lessons": len(self.lessons),
            "by_type": by_type,
            "avg_relevance": round(avg_relevance, 2)
        }
