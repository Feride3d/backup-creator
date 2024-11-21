package model

import "time"

type ContentBlock struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	ModifiedDate time.Time `json:"modifiedDate"`
	Content      string    `json:"content"`
}
