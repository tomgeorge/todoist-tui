package ctx

import (
	"github.com/tomgeorge/todoist-tui/config"
	"github.com/tomgeorge/todoist-tui/services/sync"
	"go.uber.org/zap"
)

type Context struct {
	Logger *zap.SugaredLogger
	Client *sync.Client
	Config *config.Config
}

func New(config *config.Config, logger *zap.SugaredLogger, client *sync.Client) Context {
	return Context{
		Logger: logger,
		Client: client,
		Config: config,
	}
}
