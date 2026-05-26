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
)

func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	userID, _ := utils.GetUserIDFromSession(r, h.DB)

	// Vérifie si mute
	var muted int
	h.DB.QueryRow("SELECT muted FROM users WHERE id = ?", userID).Scan(&muted)
	if muted == 1 {
		http.Error(w, "Vous êtes mute et ne pouvez pas commenter", http.StatusForbidden)
		return
	}

	// Parse multipart (20 Mo max)
	r.ParseMultipartForm(20 << 20)

	postID := r.FormValue("post_id")
	content := r.FormValue("content")

	if postID == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Gestion de l'image
	imagePath := ""
	file, header, err := r.FormFile("image")
	if err == nil {
		defer file.Close()

		ext := strings.ToLower(filepath.Ext(header.Filename))
		allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}
		if !allowed[ext] {
			http.Redirect(w, r, "/post/"+postID, http.StatusSeeOther)
			return
		}

		if header.Size > 20<<20 {
			http.Redirect(w, r, "/post/"+postID, http.StatusSeeOther)
			return
		}

		filename := fmt.Sprintf("comment_%s%s", uuid.New().String(), ext)
		dst, err := os.Create("web/static/uploads/" + filename)
		if err != nil {
			h.InternalError(w, r)
			return
		}
		defer dst.Close()
		io.Copy(dst, file)
		imagePath = filename
	}

	// Contenu ou image obligatoire
	if content == "" && imagePath == "" {
		http.Redirect(w, r, "/post/"+postID, http.StatusSeeOther)
		return
	}

	commentID := uuid.New().String()

	_, err = h.DB.Exec(
		"INSERT INTO comments (id, post_id, user_id, content, image_path) VALUES (?, ?, ?, ?, ?)",
		commentID, postID, userID, content, imagePath,
	)
	if err != nil {
		h.InternalError(w, r)
		return
	}

	http.Redirect(w, r, "/post/"+postID, http.StatusSeeOther)
}

func (h *Handler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	userID, _ := utils.GetUserIDFromSession(r, h.DB)
	commentID := r.FormValue("comment_id")
	postID := r.FormValue("post_id")

	var role string
	h.DB.QueryRow("SELECT role FROM users WHERE id = ?", userID).Scan(&role)

	if role != "admin" && role != "moderator" {
		http.Error(w, "Non autorisé", http.StatusForbidden)
		return
	}

	h.DB.Exec("DELETE FROM comments WHERE id = ?", commentID)
	http.Redirect(w, r, "/post/"+postID, http.StatusSeeOther)
}
func (h *Handler) EditComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	userID, _ := utils.GetUserIDFromSession(r, h.DB)
	commentID := r.FormValue("comment_id")
	postID := r.FormValue("post_id")
	content := r.FormValue("content")

	if content == "" {
		http.Redirect(w, r, "/post/"+postID, http.StatusSeeOther)
		return
	}

	// Vérifie que c'est bien son commentaire
	var ownerID string
	h.DB.QueryRow("SELECT user_id FROM comments WHERE id = ?", commentID).Scan(&ownerID)
	if ownerID != userID {
		http.Error(w, "Non autorisé", http.StatusForbidden)
		return
	}

	h.DB.Exec("UPDATE comments SET content = ? WHERE id = ?", content, commentID)
	http.Redirect(w, r, "/post/"+postID, http.StatusSeeOther)
}
