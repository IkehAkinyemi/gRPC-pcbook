package service

import (
	"golang.org/x/crypto/bcrypt"
)

// A User defines user's information.
type User struct {
	Username       string
	HashedPassword string
	Role           string
}

// NewUser hashes the cleartext password and returns a User instance.
func NewUser(username, password string, role string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &User{
		Username:       username,
		HashedPassword: string(hashedPassword),
		Role:           role,
	}

	return user, nil
}

// VerfiyPassword checks if the provided cleartext password
// is correct or not
func (user *User) VerfiyPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password))
	return err == nil
}

// Clone returns a clone of this user
func (user *User) Clone() *User {
	return &User{
		Username:       user.Username,
		HashedPassword: user.HashedPassword,
		Role:           user.Role,
	}
}
