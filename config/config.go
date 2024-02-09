package config

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	// Log file location
	Log string `yaml:"log"`
	// State directory
	StateDir string `yaml:"stateDir"`
	// Todoist API token
	ApiToken string `yaml:"api-token"`
}

func SetDefaults() error {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		return err
	}

	defaultLogPath := path.Join(filepath.Join(userCacheDir, "todoist-tui", "todoist-tui.log"))
	defaultStateDir := path.Join(filepath.Join(userCacheDir, "todoist-tui"))
	fmt.Println(defaultLogPath)
	viper.SetDefault("log", defaultLogPath)
	viper.SetDefault("stateDir", defaultStateDir)
	return nil
}
