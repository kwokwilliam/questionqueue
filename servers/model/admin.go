package model

import (
	"errors"
	"github.com/badoux/checkmail"
)

type Teacher struct {
	ID           int32  `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
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

func CreateNewTeacher(email, pw, pwConf, fn, ln string) (*NewTeacher, error) {

	var (
		ErrPasswordNotMatch = errors.New("passwords do not match")
		ErrEmptyName = errors.New("name cannot be empty")
	)

	if err := checkmail.ValidateFormat(email); err != nil {
		return nil, err // ErrBadFormat
	}

	if err := checkmail.ValidateHost(email); err != nil {
		return nil, err // ErrUnresolvableHost
	}

	if pw != pwConf {
	 	return nil, ErrPasswordNotMatch
	}

	if len(fn) == 0 || len(ln) == 0 {
		return nil, ErrEmptyName
	}

	return &NewTeacher{
		Email:        email,
		Password:     pw,
		PasswordConf: pwConf,
		FirstName:    fn,
		LastName:     ln,
	}, nil
}