package inventory

// Inventory represents the list of hosts to manage.
type Inventory struct {
	Hosts []Host `yaml:"hosts"`
}

// Host represents a host for automation.
// - Port is optional (defaults to 22).
// - retry_ssh: optional duration for SSH connection retry delay (e.g., "5s").
// - retry_ssh_count: optional number of SSH connection attempts.
type Host struct {
	Name          string `yaml:"name"`
	Address       string `yaml:"address"`
	Port          int    `yaml:"port,omitempty"`
	User          string `yaml:"user"`
	PrivateKey    string `yaml:"private_key,omitempty"`
	Password      string `yaml:"password,omitempty"`
	RetrySSH      string `yaml:"retry_ssh,omitempty"`
	RetrySSHCount int    `yaml:"retry_ssh_count,omitempty"`
}
