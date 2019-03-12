package handler

import (
	"encoding/json"
	"io"
	"questionqueue/src/model"
)

// decoders; probably cannot be further refactored
func decodeQuestion(d io.ReadCloser) (*model.Question, error) {
	decoder := json.NewDecoder(d)
	var i model.Question
	if err := decoder.Decode(&i); err != nil {
		return nil, err
	} else {
		return &i, nil
	}
}

func decodeQuestions(i interface{}) ([]*model.Question, error) {
	var qs []*model.Question

	// TODO: marshal list json in to list struct
}

func decodeNewTeacher(d io.ReadCloser) (*model.NewTeacher, error) {
	decoder := json.NewDecoder(d)
	var i model.NewTeacher
	if err := decoder.Decode(&i); err != nil {
		return nil, err
	} else {
		return &i, nil
	}
}

func decodeTeacher(d io.ReadCloser) (*model.Teacher, error) {
	decoder := json.NewDecoder(d)
	var i model.Teacher
	if err := decoder.Decode(&i); err != nil {
		return nil, err
	} else {
		return &i, nil
	}
}

func decodeTeacherUpdate(d io.ReadCloser) (*model.TeacherUpdate, error) {
	decoder := json.NewDecoder(d)
	var i model.TeacherUpdate
	if err := decoder.Decode(&i); err != nil {
		return nil, err
	} else {
		return &i, nil
	}
}

func decodeTeacherLogin(d io.ReadCloser) (*model.TeacherLogin, error) {
	decoder := json.NewDecoder(d)
	var i model.TeacherLogin
	if err := decoder.Decode(&i); err != nil {
		return nil, err
	} else {
		return &i, nil
	}
}
