package users

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"questionqueue/servers/gateway/indexes"
	"time"

	"github.com/go-sql-driver/mysql"
)

// GetByType is an enumerate for GetBy* functions implemented
// by MySQLStore structs
type GetByType string

// These are the enumerates for GetByType
const (
	ID       GetByType = "ID"
	Email    GetByType = "Email"
	UserName GetByType = "UserName"
)

// MySQLStore is a user.Store backed by MySQL
type MySQLStore struct {
	Database *sql.DB
}

// NewMySQLStore constructs a new MySQLStore, and returns an error
// if there is a problem along the way.
func NewMySQLStore(dataSourceName string) (*MySQLStore, error) {
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return nil, err
	}

	return &MySQLStore{db}, nil
}

// LoadUsersToTrie loads users from the database into a trie.
func (ms *MySQLStore) LoadUsersToTrie(trie *indexes.Trie) error {
	sel := "select ID, UserName, FirstName, LastName from Users"
	rows, err := ms.Database.Query(sel)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		user := &User{}
		if err := rows.Scan(
			&user.ID,
			&user.FirstName,
			&user.LastName); err != nil {
			return err
		}
		trie.AddUserToTrie(user.FirstName, user.LastName, user.ID)
	}
	return nil
}

// getByProvidedType gets a specific user given the provided type.
// This requires the GetByType to be "unique" in the database.
//
// Author question: Is factoring this an anti-pattern that is a security
//					vulnerability in cases of things like speculative execution?
func (ms *MySQLStore) getByProvidedType(t GetByType, arg interface{}) (*User, error) {
	sel := string("select ID, Email, PassHash, UserName, FirstName, LastName, PhotoURL from Users where " + t + " = ?")

	rows, err := ms.Database.Query(sel, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	user := &User{}

	// Should never have more than one row, so only grab one
	rows.Next()
	if err := rows.Scan(
		&user.ID,
		&user.Email,
		&user.PassHash,
		&user.FirstName,
		&user.LastName); err != nil {
		return nil, err
	}
	return user, nil
}

//GetByID returns the User with the given ID
func (ms *MySQLStore) GetByID(id int64) (*User, error) {
	return ms.getByProvidedType(ID, id)
}

//GetByEmail returns the User with the given email
func (ms *MySQLStore) GetByEmail(email string) (*User, error) {
	return ms.getByProvidedType(Email, email)
}

//GetByUserName returns the User with the given Username
func (ms *MySQLStore) GetByUserName(username string) (*User, error) {
	return ms.getByProvidedType(UserName, username)
}

//Insert inserts the user into the database, and returns
//the newly-inserted User, complete with the DBMS-assigned ID
func (ms *MySQLStore) Insert(user *User) (*User, error) {
	ins := "insert into Users(Email, PassHash, UserName, FirstName, LastName, PhotoURL) values (?,?,?,?,?,?)"
	res, err := ms.Database.Exec(ins, user.Email, user.PassHash,
		user.FirstName, user.LastName)
	if err != nil {
		return nil, err
	}

	lid, lidErr := res.LastInsertId()
	if lidErr != nil {
		return nil, lidErr
	}

	user.ID = lid
	return user, nil
}

// InsertSignIn inserts a sign in instance to the database and
// returns any errors that occur.
func (ms *MySQLStore) InsertSignIn(userID int64, clientIP string) error {
	ins := "insert into SuccessfulSignIns(UserID, ClientIP) values (?, ?)"
	_, err := ms.Database.Exec(ins, userID, clientIP)
	if err != nil {
		return err
	}
	return nil
}

//Update applies UserUpdates to the given user ID
//and returns the newly-updated user
func (ms *MySQLStore) Update(id int64, updates *Updates) (*User, error) {
	// Assumes updates ALWAYS includes FirstName and LastName
	upd := "update Users set FirstName = ?, LastName = ? where ID = ?"
	res, err := ms.Database.Exec(upd, updates.FirstName, updates.LastName, id)
	if err != nil {
		return nil, err
	}

	rowsAffected, rowsAffectedErr := res.RowsAffected()
	if rowsAffectedErr != nil {
		return nil, rowsAffectedErr
	}

	if rowsAffected != 1 {
		return nil, ErrUserNotFound
	}

	// Get the user using GetByID
	return ms.GetByID(id)
}

//Delete deletes the user with the given ID
func (ms *MySQLStore) Delete(id int64) error {
	del := "delete from Users where ID = ?"
	res, err := ms.Database.Exec(del, id)
	if err != nil {
		return err
	}

	rowsAffected, rowsAffectedErr := res.RowsAffected()
	if rowsAffectedErr != nil {
		return rowsAffectedErr
	}

	if rowsAffected != 1 {
		return ErrUserNotFound
	}

	return nil
}

// SetResetCode will set the reset code for the person with the specified email.
func (ms *MySQLStore) SetResetCode(email string) (string, error) {
	// Generate random hash
	randomBytes := make([]byte, 32)
	rand.Read(randomBytes)
	resetCode := base64.URLEncoding.EncodeToString(randomBytes)

	// Update query
	upd := "update Users set ResetCode=?, ResetTime=now() where Email = ?"
	res, err := ms.Database.Exec(upd, resetCode, email)
	if err != nil {
		return "", err
	}
	rowsAffected, rowsAffectedErr := res.RowsAffected()
	if rowsAffectedErr != nil {
		return "", rowsAffectedErr
	}

	if rowsAffected != 1 {
		return "", ErrUserNotFound
	}
	return resetCode, nil
}

// GetResetCodeByEmail grabs the reset code and time the reset code was created from the database
func (ms *MySQLStore) GetResetCodeByEmail(email string) (string, time.Time, error) {
	var nt mysql.NullTime
	var resetCode string
	sel := "select ResetCode, ResetTime from Users where Email = ?"
	rows, err := ms.Database.Query(sel, email)
	if err != nil {
		return "", time.Time{}, err
	}
	defer rows.Close()

	rows.Next()
	if err := rows.Scan(&resetCode, &nt); err != nil {
		return "", time.Time{}, err
	}

	if !nt.Valid {
		return "", time.Time{}, errors.New("Invalid time")
	}

	return resetCode, nt.Time, nil

}

// UpdatePassword will update the password for the specific user. I am fully
// aware this can be done in a single call alongside the select with some sort
// of sql if statement but I'm too lazy to think through that logic at this moment
func (ms *MySQLStore) UpdatePassword(email string, passHash []byte) error {
	upd := "update Users set PassHash = ?, ResetCode = NULL, ResetTime = NULL where Email = ?"
	res, err := ms.Database.Exec(upd, passHash, email)
	if err != nil {
		return err
	}
	rowsAffected, rowsAffectedErr := res.RowsAffected()
	if rowsAffectedErr != nil {
		return rowsAffectedErr
	}

	if rowsAffected != 1 {
		return ErrUserNotFound
	}

	return nil
}

// GetMultipleUsersByID gets a list of users from the user store sorted by
// ascending username
func (ms *MySQLStore) GetMultipleUsersByID(uids []int64) ([]*User, error) {
	if len(uids) == 0 {
		return []*User{}, nil
	}

	q := "(?"
	args := make([]interface{}, len(uids))
	for i := range uids {
		if i > 0 {
			q += ",?"
		}
		args[i] = uids[i]
	}
	q += ")"

	sel := "select ID, UserName, FirstName, LastName, PhotoURL from Users where ID in " + q + " order by UserName asc"

	// q, args, err := sqlx.In("select ID, UserName, FirstName, LastName, PhotoURL from Users where ID in(?) order by UserName asc", uids)
	// if err != nil {
	// 	return nil, err
	// }
	rows, err := ms.Database.Query(sel, args...)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var retUsers []*User

	for rows.Next() {
		user := &User{}
		if err := rows.Scan(
			&user.ID,
			&user.FirstName,
			&user.LastName); err != nil {
			return nil, err
		}
		retUsers = append(retUsers, user)
	}

	return retUsers, nil
}
