package factory

import (
	"github.com/google/wire"
	"github.com/ruskotwo/ready-checker/internal/domain/telegram"
)

var telegramSet = wire.NewSet(
	appSet,
	telegram.NewBot,
)
