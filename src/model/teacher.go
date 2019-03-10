package model

import (
	"errors"
	"github.com/badoux/checkmail"
	"golang.org/x/crypto/bcrypt"
)

type Teacher struct {
	ID           interface{} `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash,omitempty"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
}

type TeacherUpdate struct {
	Email        string `json:"email"`
	Password     string `json:"password"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
}

type NewTeacher struct {
	Email        string `json:"email"`
	Password     string `json:"password"`
	PasswordConf string `json:"password_conf"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
}

type TeacherLogin struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

// VerifyNewTeacher verifies `model.NewTeacher` and returns error if found any.
func (nt *NewTeacher) VerifyNewTeacher() error {

	if nt.Password != nt.PasswordConf {
		return errors.New("passwords do not match")
	}

	if err := checkmail.ValidateFormat(nt.Email); err != nil {
		return err // ErrBadFormat
	}

	if err := checkmail.ValidateHost(nt.Email); err != nil {
		return err // ErrUnresolvableHost
	}

	if len(nt.FirstName) == 0 {
		return errors.New("first name cannot be empty")
	}

	if len(nt.LastName) == 0 {
		return errors.New("last name cannot be empty")
	}

	return nil
}

// VerifyTeacher verifies `model.Teacher` and returns error if found any.
func (tu *TeacherUpdate) VerifyTeacherUpdate() error {

	if err := checkmail.ValidateFormat(tu.Email); err != nil {
		return err // ErrBadFormat
	}

	if err := checkmail.ValidateHost(tu.Email); err != nil {
		return err // ErrUnresolvableHost
	}

	if len(tu.FirstName) == 0 {
		return errors.New("first name cannot be empty")
	}

	if len(tu.LastName) == 0 {
		return errors.New("last name cannot be empty")
	}

	return nil
}

// Authenticate compares the plaintext password against the stored hash
// and returns an error if they don't match, or nil if they do
func (t *Teacher) Authenticate(password string) error {
	// use the bcrypt package to compare the supplied
	// password with the stored PassHash
	// https://godoc.org/golang.org/x/crypto/bcrypt

	if err := bcrypt.CompareHashAndPassword([]byte(t.PasswordHash), []byte(password)); err != nil {
		return err
	} else {
		return nil
	}

}