package factory

import (
	"log/slog"
	"os"

	"github.com/google/wire"
	redislib "github.com/redis/go-redis/v9"
	"github.com/ruskotwo/ready-checker/internal/config"
)

var appSet = wire.NewSet(
	provideLogger,
	config.NewConfig,
	config.NewRedisOptions,
	redislib.NewClient,
)

func provideLogger() *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)

	return slog.New(handler)
}
