"""
Hephaestus Phase Loop - Main task execution loop for agents.

Polls beads for tasks, requests file leases, executes tasks using LLM,
and updates task status.
"""

import json
import subprocess
import logging
import requests
from pathlib import Path
from typing import List, Dict, Optional, Any
from dataclasses import dataclass

from agent.llm_client import LLMClient


@dataclass
class Task:
    """Represents a beads task."""
    id: str
    title: str
    status: str
    phase: str
    description: str
    assignee: Optional[str] = None


@dataclass
class FileLease:
    """Represents a file lease from MCP."""
    lease_id: str
    file_path: str
    agent_name: str


class HephaestusLoop:
    """Main phase loop for task execution."""
    
    def __init__(
        self,
        agent_name: str,
        phases: List[str],
        llm_client: LLMClient,
        playbook,
        beads_db_path: str,
        mcp_url: str,
        heartbeat_manager,
        logger: logging.Logger
    ):
        self.agent_name = agent_name
        self.phases = [p.strip() for p in phases if p.strip()]
        self.llm_client = llm_client
        self.playbook = playbook
        self.beads_db_path = Path(beads_db_path)
        self.mcp_url = mcp_url.rstrip("/")
        self.heartbeat_manager = heartbeat_manager
        self.logger = logger
        
        self.current_task: Optional[Task] = None
        self.active_leases: List[FileLease] = []
        
        self.logger.info(f"Phase loop initialized for phases: {', '.join(self.phases)}")
    
    def iterate(self):
        """Execute one iteration of the phase loop."""
        try:
            # Poll for tasks
            task = self._poll_for_task()
            
            if not task:
                # No task available, stay idle
                self.heartbeat_manager.update_status("idle")
                return
            
            # Execute the task
            self._execute_task(task)
            
        except Exception as e:
            self.logger.error(f"Error in phase loop iteration: {e}", exc_info=True)
            self.heartbeat_manager.update_status("error", error=str(e))
    
    def _poll_for_task(self) -> Optional[Task]:
        """Poll beads for tasks matching agent phases."""
        try:
            # Use bd CLI to get tasks
            result = subprocess.run(
                ["bd", "list", "--json", "--status", "open"],
                cwd=self.beads_db_path,
                capture_output=True,
                text=True,
                timeout=10
            )
            
            if result.returncode != 0:
                self.logger.warning(f"bd list failed: {result.stderr}")
                return None
            
            # Parse tasks
            tasks_data = json.loads(result.stdout) if result.stdout.strip() else []
            
            # Filter tasks by phase
            for task_data in tasks_data:
                phase = task_data.get("phase", "").lower()
                if phase in [p.lower() for p in self.phases]:
                    return Task(
                        id=task_data.get("id", ""),
                        title=task_data.get("title", ""),
                        status=task_data.get("status", "open"),
                        phase=phase,
                        description=task_data.get("description", ""),
                        assignee=task_data.get("assignee")
                    )
            
            return None
            
        except subprocess.TimeoutExpired:
            self.logger.warning("bd list command timed out")
            return None
        except json.JSONDecodeError as e:
            self.logger.warning(f"Failed to parse bd output: {e}")
            return None
        except Exception as e:
            self.logger.error(f"Error polling for tasks: {e}", exc_info=True)
            return None
    
    def _execute_task(self, task: Task):
        """Execute a task using the LLM."""
        self.current_task = task
        self.logger.info(f"Executing task {task.id}: {task.title}")
        
        try:
            # Update task status to in_progress
            self._update_task_status(task.id, "in_progress")
            self.heartbeat_manager.update_status("working", current_task=task.id)
            
            # Request file leases
            files_to_lease = self._identify_files_for_task(task)
            leases = self._request_file_leases(files_to_lease)
            
            # Build context
            context = self._build_context(task, leases)
            
            # Load playbook lessons
            lessons = self.playbook.get_relevant_lessons(task.phase, task.description)
            
            # Generate prompt
            prompt = self._generate_prompt(task, context, lessons)
            
            # Call LLM
            self.logger.info(f"Calling LLM for task {task.id}")
            result = self.llm_client.complete(
                prompt=prompt,
                system_prompt=self._get_system_prompt(),
                max_tokens=8192,
                temperature=0.7
            )
            
            # Parse and execute action plan
            self._execute_action_plan(result.content, leases)
            
            # Update task status to complete
            self._update_task_status(task.id, "complete")
            
            # Reflect and learn
            self.playbook.reflect_on_task(task, result.content, "success")
            
            self.logger.info(f"Task {task.id} completed successfully")
            
        except Exception as e:
            self.logger.error(f"Error executing task {task.id}: {e}", exc_info=True)
            self._update_task_status(task.id, "open")  # Return to open
            self.playbook.reflect_on_task(task, "", f"error: {e}")
            
        finally:
            # Release all leases
            self._release_all_leases()
            self.current_task = None
            self.heartbeat_manager.update_status("idle")
    
    def _identify_files_for_task(self, task: Task) -> List[str]:
        """Identify files that need to be leased for the task."""
        # Simple heuristic: extract file paths from task description
        # In a real implementation, this could be more sophisticated
        files = []
        
        # Look for common file patterns in description
        import re
        patterns = [
            r'`([^`]+\.[a-z]+)`',  # Files in backticks
            r'file:\s*([^\s]+)',    # file: prefix
            r'([a-z_]+/[a-z_/]+\.[a-z]+)',  # Path-like patterns
        ]
        
        for pattern in patterns:
            matches = re.findall(pattern, task.description, re.IGNORECASE)
            files.extend(matches)
        
        # Remove duplicates
        return list(set(files))
    
    def _request_file_leases(self, files: List[str]) -> List[FileLease]:
        """Request file leases from MCP."""
        leases = []
        
        for file_path in files:
            try:
                response = requests.post(
                    f"{self.mcp_url}/leases",
                    json={
                        "file_path": file_path,
                        "agent_name": self.agent_name
                    },
                    timeout=5
                )
                
                if response.status_code == 200:
                    data = response.json()
                    lease = FileLease(
                        lease_id=data.get("lease_id", ""),
                        file_path=file_path,
                        agent_name=self.agent_name
                    )
                    leases.append(lease)
                    self.active_leases.append(lease)
                    self.logger.info(f"Acquired lease for {file_path}")
                else:
                    self.logger.warning(
                        f"Failed to acquire lease for {file_path}: {response.status_code}"
                    )
                    
            except Exception as e:
                self.logger.warning(f"Error requesting lease for {file_path}: {e}")
        
        return leases
    
    def _build_context(self, task: Task, leases: List[FileLease]) -> Dict[str, Any]:
        """Build context from task description and leased files."""
        context = {
            "task_id": task.id,
            "task_title": task.title,
            "task_description": task.description,
            "phase": task.phase,
            "files": {}
        }
        
        # Read leased files
        for lease in leases:
            try:
                file_path = self.beads_db_path / lease.file_path
                if file_path.exists():
                    with open(file_path, 'r', encoding='utf-8') as f:
                        context["files"][lease.file_path] = f.read()
            except Exception as e:
                self.logger.warning(f"Error reading file {lease.file_path}: {e}")
        
        return context
    
    def _get_system_prompt(self) -> str:
        """Get the system prompt for the LLM."""
        return """You are an AI coding agent working on software development tasks.
Your role is to analyze tasks, plan solutions, and execute file operations.

When responding, provide your action plan in the following JSON format:
{
  "analysis": "Your analysis of the task",
  "plan": ["Step 1", "Step 2", ...],
  "actions": [
    {"type": "read", "file": "path/to/file"},
    {"type": "write", "file": "path/to/file", "content": "file content"},
    {"type": "delete", "file": "path/to/file"}
  ]
}

Be thorough but concise. Focus on delivering working code."""
    
    def _generate_prompt(
        self,
        task: Task,
        context: Dict[str, Any],
        lessons: List[Dict[str, Any]]
    ) -> str:
        """Generate the prompt for the LLM."""
        prompt_parts = [
            f"# Task: {task.title}",
            f"\n## Description\n{task.description}",
            f"\n## Phase\n{task.phase}",
        ]
        
        # Add file context
        if context.get("files"):
            prompt_parts.append("\n## Current Files")
            for file_path, content in context["files"].items():
                prompt_parts.append(f"\n### {file_path}\n```\n{content}\n```")
        
        # Add playbook lessons
        if lessons:
            prompt_parts.append("\n## Relevant Lessons from Past Tasks")
            for lesson in lessons[:5]:  # Limit to top 5 lessons
                prompt_parts.append(
                    f"\n- Context: {lesson.get('context', '')}\n"
                    f"  Action: {lesson.get('action', '')}\n"
                    f"  Outcome: {lesson.get('outcome', '')}\n"
                    f"  Learned: {lesson.get('learned', '')}"
                )
        
        prompt_parts.append(
            "\n## Instructions\n"
            "Analyze the task and provide your action plan in JSON format."
        )
        
        return "\n".join(prompt_parts)
    
    def _execute_action_plan(self, llm_response: str, leases: List[FileLease]):
        """Parse LLM response and execute file operations."""
        try:
            # Try to extract JSON from response
            import re
            json_match = re.search(r'\{.*\}', llm_response, re.DOTALL)
            if not json_match:
                self.logger.warning("No JSON found in LLM response")
                return
            
            action_plan = json.loads(json_match.group())
            actions = action_plan.get("actions", [])
            
            self.logger.info(f"Executing {len(actions)} actions")
            
            for action in actions:
                action_type = action.get("type")
                file_path = action.get("file")
                
                if not file_path:
                    continue
                
                # Safety check: only operate on leased files
                leased_files = [l.file_path for l in leases]
                if file_path not in leased_files:
                    self.logger.warning(
                        f"Skipping action on non-leased file: {file_path}"
                    )
                    continue
                
                full_path = self.beads_db_path / file_path
                
                if action_type == "read":
                    # Already read in context building
                    pass
                    
                elif action_type == "write":
                    content = action.get("content", "")
                    full_path.parent.mkdir(parents=True, exist_ok=True)
                    with open(full_path, 'w', encoding='utf-8') as f:
                        f.write(content)
                    self.logger.info(f"Wrote file: {file_path}")
                    
                elif action_type == "delete":
                    if full_path.exists():
                        full_path.unlink()
                        self.logger.info(f"Deleted file: {file_path}")
                        
        except json.JSONDecodeError as e:
            self.logger.error(f"Failed to parse action plan JSON: {e}")
        except Exception as e:
            self.logger.error(f"Error executing action plan: {e}", exc_info=True)
    
    def _update_task_status(self, task_id: str, status: str):
        """Update task status in beads."""
        try:
            result = subprocess.run(
                ["bd", "update", task_id, "--status", status],
                cwd=self.beads_db_path,
                capture_output=True,
                text=True,
                timeout=10
            )
            
            if result.returncode == 0:
                self.logger.info(f"Updated task {task_id} status to {status}")
            else:
                self.logger.warning(
                    f"Failed to update task status: {result.stderr}"
                )
                
        except Exception as e:
            self.logger.error(f"Error updating task status: {e}", exc_info=True)
    
    def _release_all_leases(self):
        """Release all active file leases."""
        for lease in self.active_leases:
            try:
                response = requests.post(
                    f"{self.mcp_url}/leases/{lease.lease_id}/release",
                    timeout=5
                )
                
                if response.status_code == 200:
                    self.logger.info(f"Released lease for {lease.file_path}")
                else:
                    self.logger.warning(
                        f"Failed to release lease {lease.lease_id}: "
                        f"{response.status_code}"
                    )
                    
            except Exception as e:
                self.logger.warning(f"Error releasing lease {lease.lease_id}: {e}")
        
        self.active_leases.clear()
    
    def cleanup(self):
        """Clean up resources."""
        self.logger.info("Cleaning up phase loop")
        self._release_all_leases()
