package ctx

import (
	"github.com/tomgeorge/todoist-tui/config"
	"github.com/tomgeorge/todoist-tui/services/sync"
	"github.com/tomgeorge/todoist-tui/theme"
	"go.uber.org/zap"
)

type Context struct {
	Logger *zap.SugaredLogger
	Client *sync.Client
	Config *config.Config
	Theme  *theme.Theme
}

func New(config *config.Config, logger *zap.SugaredLogger, client *sync.Client) Context {
	return Context{
		Logger: logger,
		Client: client,
		Config: config,
		Theme:  theme.ThemeBase(),
	}
}

func (ctx Context) WithTheme(theme *theme.Theme) Context {
	ctx.Theme = theme
	return ctx
}
