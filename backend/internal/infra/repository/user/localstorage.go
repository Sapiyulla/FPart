package user

import (
	"fpart/internal/domain/user"
	"sync"
)

type UserLocalStorageRepository struct {
	storage map[string]user.User

	mu sync.Mutex
}

func NewUserLStorageRepository() *UserLocalStorageRepository {
	return &UserLocalStorageRepository{
		storage: map[string]user.User{},

		mu: sync.Mutex{},
	}
}

func (r *UserLocalStorageRepository) GetUserByID(id string) (*user.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if u, ok := r.storage[id]; !ok {
		return nil, user.ErrUserNotFound
	} else {
		return &u, nil
	}
}

func (r *UserLocalStorageRepository) AddNewUser(u *user.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.storage[u.GetID()]; ok {
		return user.ErrUserAlreadyExists
	}
	r.storage[u.GetID()] = *u
	return nil
}
