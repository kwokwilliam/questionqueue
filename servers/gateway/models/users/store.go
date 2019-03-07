package users

import (
	"errors"
	"questionqueue/servers/gateway/indexes"
)

//ErrUserNotFound is returned when the user can't be found
var ErrUserNotFound = errors.New("user not found")

//Store represents a store for Users
type Store interface {
	//GetByID returns the User with the given ID
	GetByID(id int64) (*User, error)

	//GetByEmail returns the User with the given email
	GetByEmail(email string) (*User, error)

	//GetByUserName returns the User with the given Username
	GetByUserName(username string) (*User, error)

	//Insert inserts the user into the database, and returns
	//the newly-inserted User, complete with the DBMS-assigned ID
	Insert(user *User) (*User, error)

	//Update applies UserUpdates to the given user ID
	//and returns the newly-updated user
	Update(id int64, updates *Updates) (*User, error)

	//Delete deletes the user with the given ID
	Delete(id int64) error

	// InsertSignIn inserts a sign in instance to the database and
	// returns any errors that occur.
	InsertSignIn(userID int64, clientIP string) error

	// LoadUsersToTrie loads all users in the db to the trie
	LoadUsersToTrie(trie *indexes.Trie) error

	// GetMultipleUsersByID gets a list of users from the user store sorted by
	// ascending username
	GetMultipleUsersByID(uids []int64) ([]*User, error)
}
