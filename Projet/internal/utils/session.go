package utils

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const SessionCookieName = "session_id"

func CreateSession(w http.ResponseWriter, db *sql.DB, userID string) error {
	// Supprime l'ancienne session si elle existe
	db.Exec("DELETE FROM sessions WHERE user_id = ?", userID)

	sessionID := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)

	_, err := db.Exec(
		"INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)",
		sessionID, userID, expiresAt,
	)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    sessionID,
		Expires:  expiresAt,
		HttpOnly: true,
		Path:     "/",
	})
	return nil
}

func DestroySession(w http.ResponseWriter, db *sql.DB, r *http.Request) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return
	}
	db.Exec("DELETE FROM sessions WHERE id = ?", cookie.Value)
	http.SetCookie(w, &http.Cookie{
		Name:    SessionCookieName,
		Value:   "",
		Expires: time.Unix(0, 0),
		Path:    "/",
	})
}

func GetUserIDFromSession(r *http.Request, db *sql.DB) (string, error) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return "", err
	}

	var userID string
	var expiresAt time.Time
	err = db.QueryRow(
		"SELECT user_id, expires_at FROM sessions WHERE id = ?",
		cookie.Value,
	).Scan(&userID, &expiresAt)
	if err != nil {
		return "", err
	}

	if time.Now().After(expiresAt) {
		db.Exec("DELETE FROM sessions WHERE id = ?", cookie.Value)
		return "", http.ErrNoCookie
	}

	return userID, nil
}
