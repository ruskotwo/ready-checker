package pending

import "time"

type Pending struct {
	storage *Storage
	timer   *Timer
}

func NewPending(storage *Storage, timer *Timer) *Pending {
	return &Pending{
		storage: storage,
		timer:   timer,
	}
}

func (p *Pending) Start(
	id string,
	statuses Statuses,
	duration time.Duration,
) (chan bool, time.Duration) {
	p.storage.Clean(id)
	p.storage.SetMany(id, statuses)

	return p.timer.Start(id, duration)
}

func (p *Pending) HandleStatus(
	id string,
	username string,
	status Status,
) (result Status, statuses Statuses) {
	p.storage.Set(id, username, status)

	result, statuses = p.GetStatusesWithResult(id)

	if result != Wait {
		// Останавливаем таймер, если никого больше не ждём
		p.timer.Stop(id)
	}

	return
}

func (p *Pending) GetStatusesWithResult(id string) (result Status, statuses Statuses) {
	statuses = p.storage.Get(id)

	for _, status := range statuses {
		if status == Wait {
			result = Wait
			break
		}
		if result == Undefined {
			result = status
		}
		if result != status {
			result = Undefined
		}
	}

	return
}
