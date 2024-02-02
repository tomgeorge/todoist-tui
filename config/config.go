package config

type Config struct {
	// Log file location
	Log string `yaml:"log"`
	// Todoist API token
	ApiToken string `yaml:"api-token"`
}
