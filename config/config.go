package config

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/viper"
)

type Config struct {
	// Log file location
	Log string `yaml:"log"`
	// Todoist API token
	ApiToken string `yaml:"api-token"`
}

func SetDefaults() error {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		return err
	}
	defaultLogPath := path.Join(userCacheDir, "todoist-tui.log")
	fmt.Println(defaultLogPath)
	viper.SetDefault("log", defaultLogPath)
	return nil
}
