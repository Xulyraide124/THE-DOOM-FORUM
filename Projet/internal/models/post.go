package models

import "time"

type Post struct {
	ID         string
	UserID     string
	Username   string
	Title      string
	Content    string
	ImagePath  string
	Categories []string
	Likes      int
	Dislikes   int
	UserVote   int // 1, -1 ou 0
	CreatedAt  time.Time
}
