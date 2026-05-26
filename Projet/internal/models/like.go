package models

type Like struct {
	ID         string
	UserID     string
	TargetID   string
	TargetType string // "post" ou "comment"
	Value      int    // 1 ou -1
}
