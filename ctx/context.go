package ctx

import (
	"github.com/tomgeorge/todoist-tui/config"
	"github.com/tomgeorge/todoist-tui/services/sync"
	"github.com/tomgeorge/todoist-tui/theme"
	"go.uber.org/zap"
)

// TODO: Document fields
type Context struct {
	Logger            *zap.SugaredLogger
	Client            *sync.Client
	Config            *config.Config
	Theme             *theme.Theme
	ScreenWidth       int
	ScreenHeight      int
	MainContentWidth  int
	MainContentHeight int
}

type ContextOption func(ctx *Context)

func New(config *config.Config, logger *zap.SugaredLogger, client *sync.Client, opts ...ContextOption) Context {
	ctx := &Context{
		Logger: logger,
		Client: client,
		Config: config,
		Theme:  theme.ThemeBase(),
	}

	for _, opt := range opts {
		opt(ctx)
	}
	return *ctx
}

func (ctx Context) WithTheme(theme *theme.Theme) Context {
	ctx.Theme = theme
	return ctx
}
