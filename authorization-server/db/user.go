package db

import (
	"authorization-server/model"
	"fmt"
	"github.com/google/uuid"
)

var ErrUserNotFound = fmt.Errorf("user not found")

type UserDb interface {
	Save(user model.User) (model.User, error)
	Find(username string) (model.User, error)
}

type InMemoryUserDb struct {
	users map[string]model.User
}

func NewInMemoryUserDb() *InMemoryUserDb {
	return &InMemoryUserDb{users: make(map[string]model.User)}
}

func (i *InMemoryUserDb) Save(user model.User) (model.User, error) {
	user.Id = uuid.New().String()
	i.users[user.Username] = user
	return user, nil
}

func (i *InMemoryUserDb) Find(username string) (model.User, error) {
	user, ok := i.users[username]
	if !ok {
		return user, ErrUserNotFound
	}
	return user, nil
}
