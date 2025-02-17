package pending

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strconv"
)

const (
	version = 1
)

type Storage struct {
	client *redis.Client
}

func NewStorage(client *redis.Client) *Storage {
	return &Storage{
		client,
	}
}

func (s Storage) Clean(id string) {
	s.client.Del(context.Background(), s.makeKey(id))
}

func (s Storage) Set(id string, user string, status Status) {
	s.client.HSet(context.Background(), s.makeKey(id), user, status)
}

func (s Storage) SetMany(id string, pending Statuses) {
	values := make([]interface{}, 0, len(pending))

	for user, status := range pending {
		values = append(values, user, status)
	}

	s.client.HSet(context.Background(), s.makeKey(id), values)
}

func (s Storage) Get(id string) Statuses {
	value := s.client.HGetAll(context.Background(), s.makeKey(id))

	result := make(Statuses, len(value.Val()))
	for i, v := range value.Val() {
		value, err := strconv.Atoi(v)
		if err != nil {
			value = 0
		}

		result[i] = Status(value)
	}

	return result
}

func (s Storage) makeKey(id string) string {
	return fmt.Sprintf("pending_storage_v%d_%s", version, id)
}
