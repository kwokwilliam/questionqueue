package model

import (
	"time"
)

type Question struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Class       string    `json:"class"`
	Topic       string    `json:"topic"`
	Description string    `json:"description"`
	Loc_X       float64   `json:"loc_x" bson:"loc_x"`
	Loc_Y       float64   `json:"loc_y" bson:"loc_y"`
	CreatedAt   time.Time `json:"created_at"`
}