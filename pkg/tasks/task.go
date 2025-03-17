package tasks

// Task represents an automation task defined in the YAML file.
type Task struct {
	Name       string            `yaml:"name"`
	Module     string            `yaml:"module"`
	Params     map[string]string `yaml:"params"`
	MaxRetries int               `yaml:"max_retries,omitempty"` // Maximum number of retries for this task.
	RetryDelay string            `yaml:"retry_delay,omitempty"` // Delay between retries (e.g., "3s"). Used if set.
}
