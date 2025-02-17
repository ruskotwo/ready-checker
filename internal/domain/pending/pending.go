package pending

type Pending struct {
	storage *Storage
}

func NewPending(storage *Storage) *Pending {
	return &Pending{
		storage,
	}
}

func (p Pending) Start(
	id string,
	statuses Statuses,
) error {
	p.storage.Clean(id)
	p.storage.SetMany(id, statuses)

	return nil
}

func (p Pending) Update(
	id string,
	username string,
	status Status,
) (result Status, statuses Statuses, err error) {
	p.storage.Set(id, username, status)

	statuses = p.storage.Get(id)

	statuses[username] = status

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
