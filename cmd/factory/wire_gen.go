// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package factory

import (
	"github.com/redis/go-redis/v9"
	"github.com/ruskotwo/ready-checker/internal/config"
	"github.com/ruskotwo/ready-checker/internal/domain/pending"
	"github.com/ruskotwo/ready-checker/internal/handler/telegram"
)

// Injectors from wire.go:

func InitTelegramBot() (*telegram.Bot, func(), error) {
	appConfig := config.NewConfig()
	logger := provideLogger()
	options := config.NewRedisOptions()
	client := redis.NewClient(options)
	storage := pending.NewStorage(client)
	timer := pending.NewTimer(logger)
	pendingPending := pending.NewPending(storage, timer)
	bot := telegram.NewBot(appConfig, logger, pendingPending)
	return bot, func() {
	}, nil
}
