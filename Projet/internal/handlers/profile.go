package handlers

import (
	"fmt"
	"forum/internal/utils"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) ViewProfile(w http.ResponseWriter, r *http.Request) {
	targetUsername := strings.TrimPrefix(r.URL.Path, "/user/")
	if targetUsername == "" {
		h.NotFound(w, r)
		return
	}

	var targetID, email, role, avatar, bio string
	var muted int
	err := h.DB.QueryRow(
		"SELECT id, email, role, avatar, bio, muted FROM users WHERE username = ?",
		targetUsername,
	).Scan(&targetID, &email, &role, &avatar, &bio, &muted)
	if err != nil {
		h.NotFound(w, r)
		return
	}

	// Posts de l'utilisateur
	rows, err := h.DB.Query(`
		SELECT p.id, p.title, p.created_at
		FROM posts p WHERE p.user_id = ?
		ORDER BY p.created_at DESC LIMIT 10`, targetID)
	if err != nil {
		h.InternalError(w, r)
		return
	}
	defer rows.Close()

	type PostRow struct {
		ID        string
		Title     string
		CreatedAt string
	}
	var posts []PostRow
	for rows.Next() {
		var p PostRow
		var createdAt string
		rows.Scan(&p.ID, &p.Title, &createdAt)
		if len(createdAt) >= 16 {
			p.CreatedAt = createdAt[:16]
		}
		posts = append(posts, p)
	}

	sessionUserID, _ := utils.GetUserIDFromSession(r, h.DB)
	var sessionUsername, sessionRole string
	if sessionUserID != "" {
		h.DB.QueryRow("SELECT username, role FROM users WHERE id = ?", sessionUserID).Scan(&sessionUsername, &sessionRole)
	}

	h.Templates["profile"].ExecuteTemplate(w, "base", map[string]interface{}{
		"TargetUsername": targetUsername,
		"TargetRole":     role,
		"TargetAvatar":   avatar,
		"TargetBio":      bio,
		"TargetMuted":    muted,
		"Posts":          posts,
		"IsOwner":        sessionUserID == targetID,
		"UserID":         sessionUserID,
		"Username":       sessionUsername,
		"Role":           sessionRole,
	})
}

func (h *Handler) EditProfile(w http.ResponseWriter, r *http.Request) {
	userID, _ := utils.GetUserIDFromSession(r, h.DB)

	var username, email, bio, avatar string
	h.DB.QueryRow("SELECT username, email, bio, avatar FROM users WHERE id = ?", userID).Scan(&username, &email, &bio, &avatar)

	if r.Method == http.MethodGet {
		var sessionRole string
		h.DB.QueryRow("SELECT role FROM users WHERE id = ?", userID).Scan(&sessionRole)
		h.Templates["edit_profile"].ExecuteTemplate(w, "base", map[string]interface{}{
			"Username": username,
			"Email":    email,
			"Bio":      bio,
			"Avatar":   avatar,
			"UserID":   userID,
			"Role":     sessionRole,
		})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (20 Mo max)
	r.ParseMultipartForm(20 << 20)

	newUsername := r.FormValue("username")
	newBio := r.FormValue("bio")
	newPassword := r.FormValue("password")

	if newUsername == "" {
		newUsername = username
	}

	// Upload avatar
	file, header, err := r.FormFile("avatar")
	newAvatar := avatar
	if err == nil {
		defer file.Close()

		ext := strings.ToLower(filepath.Ext(header.Filename))
		allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}
		if !allowed[ext] {
			http.Error(w, "Format non supporté (jpg, png, gif)", http.StatusBadRequest)
			return
		}

		if header.Size > 20<<20 {
			http.Error(w, "Image trop lourde (max 20 Mo)", http.StatusBadRequest)
			return
		}

		filename := fmt.Sprintf("avatar_%s%s", uuid.New().String(), ext)
		dst, err := os.Create("web/static/uploads/" + filename)
		if err != nil {
			h.InternalError(w, r)
			return
		}
		defer dst.Close()
		io.Copy(dst, file)
		newAvatar = filename
	}

	// Update username + bio + avatar
	_, err = h.DB.Exec(
		"UPDATE users SET username = ?, bio = ?, avatar = ? WHERE id = ?",
		newUsername, newBio, newAvatar, userID,
	)
	if err != nil {
		http.Error(w, "Username déjà pris", http.StatusConflict)
		return
	}

	// Update password si fourni
	if newPassword != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err == nil {
			h.DB.Exec("UPDATE users SET password = ? WHERE id = ?", string(hashed), userID)
		}
	}

	http.Redirect(w, r, "/user/"+newUsername, http.StatusSeeOther)
}
