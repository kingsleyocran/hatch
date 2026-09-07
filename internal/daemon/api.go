package daemon

type Request struct {
	Action string `json:"action"`
	Domain string `json:"domain,omitempty"`
	Port   int    `json:"port,omitempty"`
	Dir    string `json:"dir,omitempty"`
	HTTPS  bool   `json:"https,omitempty"`
}

type ProjectStatus struct {
	Domain string `json:"domain"`
	Port   int    `json:"port"`
	Dir    string `json:"dir"`
	Alive  bool   `json:"alive"`
	HTTPS  bool   `json:"https"`
}

type DaemonStatus struct {
	Running     bool   `json:"running"`
	Uptime      string `json:"uptime"`
	DomainCount int    `json:"domain_count"`
	ActiveCount int    `json:"active_count"`
}

type Response struct {
	OK       bool            `json:"ok"`
	Message  string          `json:"message,omitempty"`
	Projects []ProjectStatus `json:"projects,omitempty"`
	Status   *DaemonStatus   `json:"status,omitempty"`
}

const (
	ActionAdd    = "add"
	ActionRemove = "remove"
	ActionList   = "list"
	ActionStop   = "stop"
	ActionPing   = "ping"
	ActionStatus = "status"
)
