package models

import "time"

type User struct {
	ID        string
	Email     string
	Username  string
	Password  string
	CreatedAt time.Time
}
