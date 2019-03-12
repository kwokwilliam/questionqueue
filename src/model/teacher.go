package model

import (
	"encoding/json"
	"errors"
	"github.com/badoux/checkmail"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type Teacher struct {
	ID           primitive.ObjectID `json:"id"                      bson:"_id"`
	Email        string             `json:"email"                   bson:"email"`
	PasswordHash string             `json:"password_hash,omitempty" bson:"passwordhash"`
	FirstName    string             `json:"first_name"              bson:"firstname"`
	LastName     string             `json:"last_name"               bson:"lastname"`
}

type TeacherUpdate struct {
	Email       string `json:"email"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
}

func (tu *TeacherUpdate) ToString() string {
	b, _ := json.Marshal(tu)
	return string(b)
}

type NewTeacher struct {
	Email        string `json:"email"`
	Password     string `json:"password"`
	PasswordConf string `json:"password_conf"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
}

type TeacherLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// VerifyNewTeacher verifies `model.NewTeacher` and returns error if found any.
func (nt *NewTeacher) VerifyNewTeacher() error {

	if nt.Password != nt.PasswordConf {
		return errors.New("passwords do not match")
	}

	if len(nt.Password) < 6 {
		return errors.New("password needs to be more than 6 characters long")
	}

	if err := checkmail.ValidateFormat(nt.Email); err != nil {
		return errors.New("invalid email")
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

	if len(tu.NewPassword) < 6 {
		return errors.New("password needs to be more than 6 characters long")
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