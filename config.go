package gomigrate

import (
	"log/slog"
)

func defaultConfig() *config {
	return &config{
		Directory: defaultDirectory,
		TableName: defaultTableName,
		Logger:    slog.Default(),
	}
}

type config struct {
	Directory string
	TableName string
	Logger    *slog.Logger
}

func (c *config) apply(opts ...Option) {
	for _, opt := range opts {
		opt.apply(c)
	}
}

type Option interface {
	apply(*config)
}

type optionFunc func(*config)

func (o optionFunc) apply(c *config) {
	o(c)
}

// WithDirectory sets the directory where the migration files are loaded from.
func WithDirectory(dir string) Option {
	return optionFunc(func(c *config) {
		c.Directory = dir
	})
}

// WithTableName sets the name of the table where the schema version is stored.
func WithTableName(name string) Option {
	return optionFunc(func(c *config) {
		c.TableName = name
	})
}

// WithLogger sets the logger to be used by the migrator.
func WithLogger(l *slog.Logger) Option {
	return optionFunc(func(c *config) {
		c.Logger = l
	})
}
