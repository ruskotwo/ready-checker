//go:build wireinject
// +build wireinject

package factory

import (
	"github.com/google/wire"
	"github.com/ruskotwo/ready-checker/internal/handler/telegram"
)

func InitTelegramBot() (*telegram.Bot, func(), error) {
	panic(wire.Build(telegramSet))
}
