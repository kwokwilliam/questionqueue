package model

import (
	"errors"
	"questionqueue/servers/modelrs/model/class"
	"time"
)

type Question struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Class       string    `json:"class"`
	Topic       string    `json:"topic"`
	Description string    `json:"description"`
	X           float64   `json:"x_loc"`
	Y           float64   `json:"y_loc"`
	CreatedAt   time.Time `json:"created_at"`
	ResolvedAt  time.Time `json:"resolved_at"`
}

//type NewQuestion struct {
//	ID          string    `json:"id"`
//	Name        string    `json:"name"`
//	Class       string    `json:"class"`
//	Topic       string    `json:"topic"`
//	Description string    `json:"description"`
//	X           float64   `json:"x_loc"`
//	Y           float64   `json:"y_loc"`
//	CreatedAt   time.Time `json:"created_at"`
//}

func CreateNewQuestion(id, name, class, topic, description string, x, y float64) (*Question, error) {

	var (
		ErrEmptyQuestion    = errors.New("question cannot be empty")
		ErrInvalidClass     = errors.New("invalid class")
		ErrEmptyTopic       = errors.New("topic cannot be empty")
		ErrEmptyLoc         = errors.New("location cannot be empty")
		ErrEmptyDescription = errors.New("description cannot be empty")
	)

	if len(name) == 0 {
		return nil, ErrEmptyQuestion
	}

	if !model.ValidateClass(class) {
		return nil, ErrInvalidClass
	}

	if len(topic) == 0 {
		return nil, ErrEmptyTopic
	}

	if len(description) == 0 {
		return nil, ErrEmptyDescription
	}

	if x == 0.0 && y == 0.0 {
		return nil, ErrEmptyLoc
	}

	return &Question{
		ID:          id,
		Name:        name,
		Class:       class,
		Topic:       topic,
		Description: description,
		X:           x,
		Y:           y,
		CreatedAt:   time.Now(),
		ResolvedAt:  nil,
	}, nil
}

func QuestionResolved(q *Question) bool {
	return q.ResolvedAt == nil
}