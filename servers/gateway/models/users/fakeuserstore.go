package users

import (
	"assignments-kwokwilliam/servers/gateway/indexes"
	"errors"
	"time"
)

// FakeMySQLStore is used for tests
type FakeMySQLStore struct {
}

// NewFakeStore creates a new FakeMySQLStore
func NewFakeStore() *FakeMySQLStore {
	return &FakeMySQLStore{}
}

// GetByID is a fake function for FakeMySQLStores
func (s *FakeMySQLStore) GetByID(id int64) (*User, error) {
	if id == 1 {
		return &User{
			ID: 1,
		}, nil
	}
	return nil, errors.New("only id 1 can be found")

}

// GetByEmail is a fake function for FakeMySQLStores
func (s *FakeMySQLStore) GetByEmail(email string) (*User, error) {
	if email == "emailcannotfound@email.com" {
		return nil, errors.New("Email cannot be found")
	}

	if email == "emailfailedcreds@email.com" {
		return &User{ID: 1, PassHash: []byte("not the right passcode")}, nil
	}

	if email == "emailworkingcreds@email.com" {
		user := &User{ID: 1}
		user.SetPassword("password")
		return user, nil
	}
	return &User{
		ID: 1,
	}, nil
}

// GetByUserName is a fake function for FakeMySQLStores
func (s *FakeMySQLStore) GetByUserName(username string) (*User, error) {
	return &User{
		ID: 1,
	}, nil
}

// Insert is a fake function for FakeMySQLStores
func (s *FakeMySQLStore) Insert(user *User) (*User, error) {
	return &User{
		ID: 1,
	}, nil
}

// InsertSignIn is a fake function for FakeMySQLStores
func (s *FakeMySQLStore) InsertSignIn(userID int64, clientIP string) error {
	return nil
}

// Update is a fake function for FakeMySQLStores
func (s *FakeMySQLStore) Update(id int64, updates *Updates) (*User, error) {
	return &User{
		ID: 1,
	}, nil
}

// Delete is a fake function for FakeMySQLStores
func (s *FakeMySQLStore) Delete(id int64) error { return nil }

// UpdateImage is a fake function for FakeMySQLStores
func (s *FakeMySQLStore) UpdateImage(id int64, updates *Updates) (*User, error) { return nil, nil }

// SetResetCode will set the reset code for the person with the specified email.
func (s *FakeMySQLStore) SetResetCode(email string) (string, error) { return "", nil }

// UpdatePassword sets the password again
func (s *FakeMySQLStore) UpdatePassword(email string, passHash []byte) error {
	return errors.New("Not implemented")
}

// GetResetCodeByEmail will get the user's reset code and time with the provided email
func (s *FakeMySQLStore) GetResetCodeByEmail(email string) (string, time.Time, error) {
	return "", time.Time{}, errors.New("Not implemented")
}

// LoadUsersToTrie is not implemented in fake store
func (s *FakeMySQLStore) LoadUsersToTrie(trie *indexes.Trie) error {
	return nil
}

// GetMultipleUsersByID gets a list of users from the user store sorted by
// ascending username
func (s *FakeMySQLStore) GetMultipleUsersByID(uids []int64) ([]*User, error) {
	return nil, nil
}
