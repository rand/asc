package tui

import (
	"github.com/yourusername/asc/internal/config"
	"github.com/yourusername/asc/internal/process"
)

// processManagerAdapter adapts process.ProcessManager to config.ProcessManager
type processManagerAdapter struct {
	pm process.ProcessManager
}

func newProcessManagerAdapter(pm process.ProcessManager) config.ProcessManager {
	return &processManagerAdapter{pm: pm}
}

func (a *processManagerAdapter) Start(name, command string, args []string, env []string) (int, error) {
	return a.pm.Start(name, command, args, env)
}

func (a *processManagerAdapter) Stop(pid int) error {
	return a.pm.Stop(pid)
}

func (a *processManagerAdapter) IsRunning(pid int) bool {
	return a.pm.IsRunning(pid)
}

func (a *processManagerAdapter) GetProcessInfo(name string) (config.ProcessInfoGetter, error) {
	info, err := a.pm.GetProcessInfo(name)
	if err != nil {
		return nil, err
	}
	// process.ProcessInfo already implements config.ProcessInfoGetter via its getter methods
	return info, nil
}
