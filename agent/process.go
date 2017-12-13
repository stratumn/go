package agent

// ActionInfo is the data structure used by Process.GetInfo()
type ActionInfo map[string][]string

// ProcessInfo is the data structure used to store
// information about a process actions and plugins
type ProcessInfo struct {
	Actions     map[string]ActionInfo `json:"actions"`
	PluginsInfo []PluginInfo          `json:"pluginsInfo"`
}

// Process is the agent's representation of a process
type Process struct {
	Name            string           `json:"name"`
	ProcessInfo     ProcessInfo      `json:"processInfo"`
	StoreInfo       StoreInfo        `json:"storeInfo"`
	FossilizersInfo []FossilizerInfo `json:"fossilizersInfo"`
}

// Processes is a list of Process
type Processes []*Process

// ProcessesMap is a mapping of processes indexed by name
type ProcessesMap map[string]*Process

// FindProcess returns the process whose name matches the provided one
func (p Processes) FindProcess(name string) *Process {
	for _, process := range p {
		if process.Name == name {
			return process
		}
	}
	return nil
}
