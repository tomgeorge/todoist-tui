package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tomgeorge/todoist-tui/config"
	"github.com/tomgeorge/todoist-tui/ctx"
	"github.com/tomgeorge/todoist-tui/model"
	"github.com/tomgeorge/todoist-tui/services/sync"
	"github.com/tomgeorge/todoist-tui/theme"
	"go.uber.org/zap"
)

var CFG *config.Config

type rootFlags struct {
	apiToken   string
	configFile string
}

func NewRootCmd() *cobra.Command {
	flags := rootFlags{}
	cmd := &cobra.Command{
		Use:   "todoist-tui",
		Short: "Todoist terminal UI",
		Long:  "Fill me in",
		RunE:  rootCmd,
	}
	initializeCobra(cmd)

	cmd.Flags().StringVar(&flags.apiToken, "api-token", "", "Todoist API Token")
	cmd.Flags().StringVar(&flags.configFile, "config", "", "Config file")
	viper.BindPFlags(cmd.Flags())

	return cmd
}

func initializeCobra(cmd *cobra.Command) {
	cfgFile, _ := cmd.Flags().GetString("config")
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName("todoist-tui")

		viper.SetEnvPrefix("todoist_tui")
		viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
		viper.AutomaticEnv()

		if err := viper.ReadInConfig(); err == nil {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
		config.SetDefaults()
	}
}

func loadLogger(logfilePath string) (*zap.Logger, error) {
	cfg := zap.NewDevelopmentConfig()
	cfg.OutputPaths = []string{
		logfilePath,
	}
	return cfg.Build()
}

func rootCmd(cmd *cobra.Command, args []string) error {
	var config config.Config

	err := viper.Unmarshal(&config)
	if err != nil {
		return fmt.Errorf("unmarshal config %v", err)
	}

	logDir := filepath.Dir(config.Log)
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		if err = os.MkdirAll(logDir, 0755); err != nil {
			return err
		}
	}

	logger, err := loadLogger(config.Log)
	if err != nil {
		return fmt.Errorf("loading logger: %v", err)
	}
	client := sync.NewClient(nil).WithAuthToken(viper.GetString("api-token"))
	ctx := ctx.New(&config, logger.Sugar(), client).WithTheme(theme.ThemeCatppuccin())
	model := model.New(ctx)
	p := tea.NewProgram(model)

	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
