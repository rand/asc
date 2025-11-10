# **Specification: asc (Agent Stack Controller)**

## **1\. Overview & Vision**

**Vision:** To provide a seamless, "it-just-works" developer experience for managing a local colony of AI coding agents. The asc tool is the developer's "mission control." It handles the orchestration of starting, monitoring, and coordinating the headless background agents, freeing the developer to act as a high-level "director" from their own interactive IDE sessions.

**Core User Journey:**

1. The developer types asc up in a terminal.  
2. asc verifies all dependencies, launches the mcp\_agent\_mail server, and activates all defined "headless" agents (e.g., claude-refactor, gemini-planner) as background processes.  
3. The developer's terminal transforms into a clean, succinct TUI dashboard, showing the real-time status of all agents, the shared task list (beads), and the agent-to-agent communication log.  
4. The developer opens a *new* terminal, uses their IDE (e.g., "Claude Code") to create tasks in the beads database (e.g., bd new "Refactor auth.py").  
5. The developer watches on the asc up dashboard as the headless agents pick up, coordinate, and execute these tasks.  
6. The developer can "jump in" at any time, acting as a human peer to the agent colony, by claiming a beads task themselves.

## **2\. Core Principles & Philosophy**

This system's design must be guided by the following principles:

* **DX-First (The Vercel Aesthetic):** The tool must be fast, intuitive, and require minimal configuration. It should feel powerful, not complex. Commands should be simple (up, down, check). Errors should be clear and actionable.  
* **Elegant Information Density (The Charm.land Aesthetic):** The TUI dashboard is the primary interface. It must be built using the **Go \+ Charm.land** stack (bubbletea, lipgloss, bubbles). It must present a high density of information (statuses, logs, tasks) in a way that is beautiful, responsive, and "glanceable."  
* **Zero-Framework Cognition (ZFC):** The asc tool itself, and the "adapters" it launches, must be "thin shells." All complex reasoning is delegated to the LLMs (Claude, Gemini, Codex). The code we build is for *orchestration*, not *cognition*.  
* **Peers, Not Pets:** The system's architecture (built on beads and mcp\_agent\_mail) must treat the human developer and the AI agents as peers. They share the same task database and the same communication bus. This is the core of the collaborative experience.  
* **Matrix Architecture:** The system must decouple **Agent Role** (e.g., planner, tester) from the **LLM Model** (e.g., claude, gemini). The configuration must allow any model to be assigned to any role, enabling easy mixing, matching, and comparison.

## **3\. System Architecture**

asc orchestrates a 5-layer stack:

* **L0: State & Task (beads):** The single source of truth for "what needs to be done." All tasks, dependencies, and statuses are stored in this Git-backed database.  
* **L1: Communication (mcp\_agent\_mail):** The "email" server for agents. Used for asynchronous coordination, status updates, and (most importantly) advisory file "leasing" to prevent conflicts.  
* **L2: Workflow (Hephaestus):** A *pattern* for the agents. Agents operate in "phases" (Plan, Implement, Test) and can dynamically create new beads tasks for other agents, allowing the workflow to "build itself."  
* **L3: Learning (ACE):** A *pattern* for the agent adapters. Each agent's adapter should include a Reflector and Curator loop to learn from its failures and update a "playbook," which is fed back into its context on future runs.  
* **L4: Orchestration (asc):** This tool. The Go binary that launches, manages, and monitors L1-L3 and the agents that use them.

## **4\. CLI Command Specification (asc)**

The asc tool shall be a single, static Go binary. It will be built using the cobra library for command-line parsing.

### **asc init**

* **Action:** Runs an interactive, first-time setup wizard (using bubbletea).  
* **Steps:**  
  1. Welcomes the user.  
  2. Runs asc check to identify existing tools and missing dependencies.  
  3. Offers to install missing components (e.g., git clone beads, mcp\_agent\_mail).  
  4. Searches for existing configs (e.g., \~/.beads\_config) and offers to back them up (\~/.asc\_backup/YYYY-MM-DD).  
  5. Prompts for API keys (Claude, OpenAI, Google) and securely stores them in a project-local .env file.  
  6. Creates a default asc.toml configuration file (see below).  
  7. Runs asc test to confirm a working stack.  
  8. On success, reports: "✅ Your agent stack is configured and ready. Run asc up to start."

### **asc up**

* **Action:** The primary command. Launches all services and the TUI dashboard.  
* **Steps:**  
  1. Runs asc check silently. If it fails, prints a clear error and exits (e.g., "Error: beads binary not found. Run asc init.").  
  2. Reads the asc.toml file to find defined agents.  
  3. Calls asc services start mcp\_agent\_mail to start the mail server.  
  4. For each agent in asc.toml (e.g., \[agent.main-planner\]):  
     a. Launches the defined command (e.g., python agent\_adapter.py) as a durable, "headless" background process (e.g., using os/exec and managing the Cmd struct).  
     b. Passes required environment variables (API keys from .env, MCP\_MAIL\_URL, BEADS\_DB\_PATH) and passes the agent's specific config (e.g., AGENT\_NAME=main-planner, AGENT\_MODEL=gemini, AGENT\_PHASES=planning,design) as environment variables.  
  5. Once all processes are launched, the current terminal is cleared and the bubbletea TUI application is launched, taking over the session.  
  6. The TUI will then connect to the beads DB and mcp\_agent\_mail server to populate its state.

### **asc down**

* **Action:** Gracefully shuts down all agents and services.  
* **Steps:**  
  1. Reads the asc.toml to find all managed agent processes.  
  2. Sends a SIGTERM to all child agent processes.  
  3. Calls asc services stop mcp\_agent\_mail.  
  4. Prints a "✅ Agent stack offline." message.

### **asc check**

* **Action:** A non-interactive "pre-flight" check.  
* **Steps:**  
  1. Verifies all required binaries are in the PATH: git, python3, uv, bd, docker (if required by asc.toml agents).  
  2. Verifies the asc.toml file exists and is valid.  
  3. Verifies the .env file exists and required API keys are set.  
  4. Prints a list of checks and their status (e.g., \[✅\] git, \[❌\] bd).  
  5. Exits with code 0 on success, 1 on failure.

### **asc test**

* **Action:** Runs a "tracer bullet" end-to-end test of the full stack *without* a TUI.  
* **Steps:**  
  1. Creates a test beads issue: "asc test task".  
  2. Sends a test mcp\_agent\_mail message.  
  3. Polls beads and mcp\_agent\_mail to confirm receipt.  
  4. Cleans up by deleting the test task and message.  
  5. Reports "✅ Stack is healthy" or "❌ Test failed at step X."

### **asc services \[start|stop|status\]**

* **Action:** A utility for managing long-running, non-agent dependencies, primarily mcp\_agent\_mail. Uses a process manager or .pid files.

## **5\. TUI Dashboard Specification (for asc up)**

This is the most critical UI component. It **must** be built with go and bubbletea.

### **Layout (using bubbletea)**

The UI will be divided into three main panes, with a footer for keybindings. lipgloss shall be used for all styling (borders, colors) to emulate the Vercel/Charm aesthetic.

\+---------------------------+--------------------------------+  
| \[1\] AGENT STATUS          | \[2\] BEADS TASK STREAM          |  
|                           |                                |  
|  \[●\] main-planner (Idle)  |  \#123 \[⟳\] Implement auth      |  
|  \[⟳\] claude-refactor (Busy)|  \#124 \[ \] Write tests for auth |  
|  \[\!\] codex-tester (Error) |  \#125 \[ \] Design new UI        |  
|                           |                                |  
\+---------------------------+--------------------------------+  
| \[3\] MCP INTERACTION LOG                                    |  
|                                                            |  
| 19:34:01 \[Lease\] claude-refactor → LEASING src/auth.py       |  
| 19:34:03 \[Beads\] main-planner → NEW-TASK \#124              |  
| 19:34:05 \[Error\] codex-tester → Test run failed: ...       |  
|                                                            |  
\+------------------------------------------------------------+  
| (q)uit | (r)efresh all | (t)rigger test                     |  
\+------------------------------------------------------------+

### **Pane 1: Agent Status**

* **Data Source:** The asc.toml file (for the list of agents) and the mcp\_agent\_mail server (for "heartbeat" messages).  
* **Function:** Polls agents for health. Displays:  
  * **Icon:** \[●\] (Idle), \[⟳\] (Working), \[\!\] (Error), \[○\] (Offline).  
  * **Name:** main-planner.  
  * **Status:** (Idle) or (Working on \#123).  
* **Styling:** Use lipgloss. Green for Idle, Blue for Working, Red for Error.

### **Pane 2: Beads Task Stream**

* **Data Source:** The local beads Git database (.beads/db.jsonl). The TUI will periodically git pull the repo and re-read the file to refresh.  
* **Function:** Displays a combined list of beads tasks with statuses open or in\_progress.  
* **Columns:** Task ID, Status Icon, Title.  
* **Styling:** Tasks in\_progress should be highlighted.

### **Pane 3: MCP Interaction Log**

* **Data Source:** The mcp\_agent\_mail server's message API.  
* **Function:** Polls the server for new messages and displays them in a tailing log.  
* **Format:** \[Timestamp\] \[Type\] \[Source\] → \[Message\].  
* **Styling:** Color-code message types:  
  * \[Lease\] (Blue)  
  * \[Beads\] (Green)  
  * \[Error\] (Red)  
  * \[Message\] (Default)

### **Keybindings (Footer)**

* (q): Quit. Triggers asc down.  
* (r): Force-refresh all panes.  
* (t): Manually trigger asc test.

## **6\. Core Component Definitions**

### **asc.toml (Configuration File)**

This TOML file defines the agent stack. It explicitly decouples the agent's name (e.g., main-planner) from its underlying model and phases (role).

\# asc.toml  
\#  
\# Agent Stack Controller Configuration  
\# This file defines the "agent colony" that \`asc up\` will launch.

\[core\]  
\# Path to the beads database (can be relative)  
beads\_db\_path \= "./project-repo"

\[services.mcp\_agent\_mail\]  
\# Command to start the mail server  
start\_command \= "python \-m mcp\_agent\_mail.server"  
\# URL for agents to connect to  
url \= "http://localhost:8765"

\# \--- Agent Definitions \---  
\#  
\# Define your headless agents here. \`asc up\` will launch one  
\# process for each agent defined.  
\#  
\# The agent's KEY (e.g., "main-planner") is its unique AGENT\_NAME.  
\#  
\# \- \`command\`: The script that implements the agent logic.  
\# \- \`model\`: The LLM to use (e.g., 'gemini', 'claude', 'codex').  
\#            The \`agent\_adapter.py\` script will use this to  
\#            instantiate the correct LLM client.  
\# \- \`phases\`: The Hephaestus-style role(s) this agent will poll for.

\[agent.main-planner\]  
command \= "python agent\_adapter.py"  
model \= "gemini"  
phases \= \["planning", "design"\]

\[agent.claude-refactor-bot\]  
command \= "python agent\_adapter.py"  
model \= "claude"  
phases \= \["implementation", "refactor"\]

\[agent.codex-test-writer\]  
command \= "python agent\_adapter.py"  
model \= "codex"  
phases \= \["testing", "validation"\]

\# \--- Specialized Agent Types \---

\[agent.high-level-orchestrator\]  
command \= "python agent\_adapter.py"  
model \= "claude" \# Use a powerful model for coordination  
phases \= \["orchestration"\] \# Listens to all \`beads\` and \`mcp\` events

\[agent.ace-optimizer\]  
command \= "python agent\_adapter.py"  
model \= "gemini"  
phases \= \["optimization"\] \# Listens for 'Reflect' and 'Curate' events

\[agent.claude-reviewer\]  
command \= "python agent\_adapter.py"  
model \= "claude"  
phases \= \["review"\] \# Can be a dependency for \`implementation\` tasks

\[agent.gemini-executor\]  
command \= "python agent\_adapter.py"  
model \= "gemini"  
phases \= \["execution"\] \# General-purpose work execution

\[agent.claude-planner\]  
\# Example of a \*second\* planner, using a different model  
\# This allows you to A/B test agent performance.  
command \= "python agent\_adapter.py"  
model \= "claude"  
phases \= \["planning"\]

### **agent\_adapter.py (The Headless Agent)**

This is the Python script that asc up launches for each agent. It is *not* part of the asc Go binary, but it is *required* for the system to function. This spec assumes asc can find it.

The adapter must:

1. Read its configuration from environment variables: AGENT\_NAME (e.g., 'main-planner'), AGENT\_MODEL (e.g., 'gemini'), and AGENT\_PHASES (e.g., 'planning,design').  
2. Based on the AGENT\_MODEL variable, it will instantiate the correct LLM client (e.g., GeminiClient, ClaudeClient, CodexClient).  
3. Based on the AGENT\_PHASES (which define its **role**), it will determine its core logic loop. For example:  
   * A planning agent's loop will poll beads for planning tasks.  
   * An orchestration agent's loop might listen to *all* beads and mcp events to manage dependencies.  
   * An optimization agent's loop might look for "failure" events and trigger an ACE reflection loop.  
4. Enter its main loop (the "Hephaestus Phase Loop").  
5. Inside the loop (ZFC):  
   a. Find Task: Call bd ready \--json \--phase {self.phases} to find an available task.  
   b. Get Lease: Send a lease message via mcp\_agent\_mail for the task's files.  
   c. Build Context (ACE): Load the "playbook" of lessons learned.  
   d. Call LLM: Send the task, context, and playbook to its instantiated LLM client and ask for a "plan" or "code."  
   e. Execute: Run the plan (e.g., write files, run tests).  
   f. Reflect (ACE): On success or failure, ask the LLM to reflect on the outcome.  
   g. Curate (ACE): If reflection generated a new "lesson," save it to the playbook.  
   h. Update State: Update beads (bd update ... \--status done) and release the mcp\_agent\_mail lease.

## **7\. Key References**

* **Aesthetics:** [Vercel](https://vercel.com), [Charm.land](https://charm.land/)  
* **TUI:** [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Lip Gloss](https://github.com/charmbracelet/lipgloss)  
* **CLI:** [Go](https://go.dev/), [Cobra](https://github.com/spf13/cobra)  
* **State:** [steveyegge/beads](https://github.com/steveyegge/beads)  
* **Communication:** [Dicklesworthstone/mcp\_agent\_mail](https://github.com/Dicklesworthstone/mcp_agent_mail)  
* **Workflow:** [Ido-Levi/Hephaestus](https://github.com/Ido-Levi/Hephaestus)  
* **Principle:** [Zero-Framework Cognition](https://steve-yegge.medium.com/zero-framework-cognition-a-way-to-build-resilient-ai-applications-56b090ed3e69)  
* **Learning:** [Agentic Context Engineering (arXiv:2510.04618)](https://www.google.com/search?q=httpsias://arxiv.org/abs/2510.04618)

## **8\. Implementation Epic (for the Agentic System)**

**Epic:** Build the asc (Agent Stack Controller)

**Tasks:**

1. **task:cli/scaffold:** Initialize a new Go project. Use cobra to scaffold the CLI commands: init, up, down, check, test, services.  
2. **task:config/toml:** Implement the asc.toml configuration file parsing (using viper or toml-go), making sure to read the \[agent.\*\] map.  
3. **task:cli/check:** Implement the full logic for the asc check command.  
4. **task:cli/init:** Implement the interactive asc init wizard using bubbletea.  
5. **task:process/manager:** Implement the core logic for asc up and asc down to launch and terminate background processes defined in asc.toml, passing the correct AGENT\_NAME, AGENT\_MODEL, and AGENT\_PHASES as env vars.  
6. **task:tui/layout:** Implement the main asc up TUI bubbletea application, focusing on the 3-pane layout using lipgloss.  
7. **task:tui/beads:** Wire the \[ Beads Task Stream \] pane. Implement the logic to poll the beads Git repo file, parse it, and display tasks.  
8. **task:tD/mcp:** Wire the \[ MCP Interaction Log \] pane. Implement the HTTP client to poll the mcp\_agent\_mail server and display messages.  
9. **task:tD/agents:** Wire the \[ Agent Status \] pane. Implement the logic to show status based on asc.toml (e.g., main-planner) and (later) health pings.  
10. **task:cli/test:** Implement the asc test "tracer bullet" logic.  
11. **task:doc/readme:** Write a README.md with installation and usage instructions.