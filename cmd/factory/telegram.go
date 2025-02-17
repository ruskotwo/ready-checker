package factory

import (
	"github.com/google/wire"
	"github.com/ruskotwo/ready-checker/internal/handler/telegram"
)

var telegramSet = wire.NewSet(
	pendingSet,
	telegram.NewBot,
)
