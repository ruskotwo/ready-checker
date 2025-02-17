package factory

import (
	"github.com/google/wire"
	"github.com/ruskotwo/ready-checker/internal/domain/pending"
)

var pendingSet = wire.NewSet(
	appSet,
	pending.NewStorage,
	pending.NewTimer,
	pending.NewPending,
)
