package reposytory

import "my-bank-service/internal/data"

// UserRepository is an interface for the storage implementation of the auth service
type UserRepository interface {
	Create(user *data.User) error
	GetUserByEmail(email string) (*data.User, error)
	GetUserByUserName(userName string) (*data.User, error)
	GetUserByID(userID string) (*data.User, error)
}
