package models

import "time"

type Comment struct {
	ID        string
	PostID    string
	UserID    string
	Username  string
	UserRole  string
	Content   string
	ImagePath string
	Likes     int
	Dislikes  int
	UserVote  int
	CreatedAt time.Time
}
