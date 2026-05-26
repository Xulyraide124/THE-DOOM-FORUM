package handlers

import (
	"database/sql"
	"fmt"
	"forum/internal/models"
	"forum/internal/utils"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		h.NotFound(w, r)
		return
	}

	userID, _ := utils.GetUserIDFromSession(r, h.DB)
	filter := r.URL.Query().Get("filter")
	category := r.URL.Query().Get("category")

	query := `
		SELECT DISTINCT p.id, p.user_id, u.username, p.title, p.content, p.image_path, p.created_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
	`
	args := []interface{}{}

	if filter == "my" && userID != "" {
		query += " WHERE p.user_id = ?"
		args = append(args, userID)
	} else if filter == "liked" && userID != "" {
		query += " JOIN likes l ON p.id = l.target_id AND l.target_type = 'post' WHERE l.user_id = ? AND l.value = 1"
		args = append(args, userID)
	} else if category != "" {
		query += " JOIN post_categories pc ON p.id = pc.post_id JOIN categories c ON pc.category_id = c.id WHERE c.name = ?"
		args = append(args, category)
	}

	query += " ORDER BY p.created_at DESC"

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		h.InternalError(w, r)
		return
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var p models.Post
		if err := rows.Scan(&p.ID, &p.UserID, &p.Username, &p.Title, &p.Content, &p.ImagePath, &p.CreatedAt); err != nil {
			continue
		}
		p.Likes, p.Dislikes = getLikeCount(h.DB, p.ID, "post")
		p.Categories = getPostCategories(h.DB, p.ID)
		posts = append(posts, p)
	}

	categories := getAllCategories(h.DB)

	var username, role string
	if userID != "" {
		h.DB.QueryRow("SELECT username, role FROM users WHERE id = ?", userID).Scan(&username, &role)
	}

	h.Templates["index"].ExecuteTemplate(w, "base", map[string]interface{}{
		"Posts":      posts,
		"Categories": categories,
		"UserID":     userID,
		"Username":   username,
		"Role":       role,
		"Filter":     filter,
		"Category":   category,
	})
}

func (h *Handler) ViewPost(w http.ResponseWriter, r *http.Request) {
	postID := strings.TrimPrefix(r.URL.Path, "/post/")
	if postID == "" {
		h.NotFound(w, r)
		return
	}

	userID, _ := utils.GetUserIDFromSession(r, h.DB)

	var p models.Post
	err := h.DB.QueryRow(`
		SELECT p.id, p.user_id, u.username, p.title, p.content, p.image_path, p.created_at
		FROM posts p JOIN users u ON p.user_id = u.id
		WHERE p.id = ?`, postID,
	).Scan(&p.ID, &p.UserID, &p.Username, &p.Title, &p.Content, &p.ImagePath, &p.CreatedAt)
	if err != nil {
		h.NotFound(w, r)
		return
	}

	p.Likes, p.Dislikes = getLikeCount(h.DB, p.ID, "post")
	p.Categories = getPostCategories(h.DB, p.ID)
	if userID != "" {
		p.UserVote = getUserVote(h.DB, userID, p.ID)
	}

	rows, err := h.DB.Query(`
		SELECT c.id, c.post_id, c.user_id, u.username, u.role, c.content, c.image_path, c.created_at
		FROM comments c JOIN users u ON c.user_id = u.id
		WHERE c.post_id = ? ORDER BY c.created_at ASC`, postID,
	)
	if err != nil {
		h.InternalError(w, r)
		return
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var c models.Comment
		if err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.Username, &c.UserRole, &c.Content, &c.ImagePath, &c.CreatedAt); err != nil {
			continue
		}
		c.Likes, c.Dislikes = getLikeCount(h.DB, c.ID, "comment")
		if userID != "" {
			c.UserVote = getUserVote(h.DB, userID, c.ID)
		}
		comments = append(comments, c)
	}

	var username, role string
	if userID != "" {
		h.DB.QueryRow("SELECT username, role FROM users WHERE id = ?", userID).Scan(&username, &role)
	}

	h.Templates["post"].ExecuteTemplate(w, "base", map[string]interface{}{
		"Post":     p,
		"Comments": comments,
		"UserID":   userID,
		"Username": username,
		"Role":     role,
	})
}

func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		categories := getAllCategories(h.DB)
		userID, _ := utils.GetUserIDFromSession(r, h.DB)
		var username, role string
		if userID != "" {
			h.DB.QueryRow("SELECT username, role FROM users WHERE id = ?", userID).Scan(&username, &role)
		}
		h.Templates["create_post"].ExecuteTemplate(w, "base", map[string]interface{}{
			"Categories": categories,
			"UserID":     userID,
			"Username":   username,
			"Role":       role,
		})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	userID, _ := utils.GetUserIDFromSession(r, h.DB)

	var muted int
	h.DB.QueryRow("SELECT muted FROM users WHERE id = ?", userID).Scan(&muted)
	if muted == 1 {
		http.Error(w, "Vous êtes mute et ne pouvez pas poster", http.StatusForbidden)
		return
	}

	r.ParseMultipartForm(20 << 20)

	title := r.FormValue("title")
	content := r.FormValue("content")
	categoryIDs := r.Form["categories"]

	if title == "" || content == "" || len(categoryIDs) == 0 {
		categories := getAllCategories(h.DB)
		var username, role string
		h.DB.QueryRow("SELECT username, role FROM users WHERE id = ?", userID).Scan(&username, &role)
		h.Templates["create_post"].ExecuteTemplate(w, "base", map[string]interface{}{
			"Error":      "Titre, contenu et au moins une catégorie sont obligatoires",
			"Categories": categories,
			"UserID":     userID,
			"Username":   username,
			"Role":       role,
		})
		return
	}

	imagePath := ""
	file, header, err := r.FormFile("image")
	if err == nil {
		defer file.Close()

		ext := strings.ToLower(filepath.Ext(header.Filename))
		allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true}
		if !allowed[ext] {
			categories := getAllCategories(h.DB)
			var username, role string
			h.DB.QueryRow("SELECT username, role FROM users WHERE id = ?", userID).Scan(&username, &role)
			h.Templates["create_post"].ExecuteTemplate(w, "base", map[string]interface{}{
				"Error":      "Format non supporté (jpg, png, gif uniquement)",
				"Categories": categories,
				"UserID":     userID,
				"Username":   username,
				"Role":       role,
			})
			return
		}

		if header.Size > 20<<20 {
			categories := getAllCategories(h.DB)
			var username, role string
			h.DB.QueryRow("SELECT username, role FROM users WHERE id = ?", userID).Scan(&username, &role)
			h.Templates["create_post"].ExecuteTemplate(w, "base", map[string]interface{}{
				"Error":      "Image trop lourde (max 20 Mo)",
				"Categories": categories,
				"UserID":     userID,
				"Username":   username,
				"Role":       role,
			})
			return
		}

		filename := fmt.Sprintf("post_%s%s", uuid.New().String(), ext)
		dst, err := os.Create("web/static/uploads/" + filename)
		if err != nil {
			h.InternalError(w, r)
			return
		}
		defer dst.Close()
		io.Copy(dst, file)
		imagePath = filename
	}

	postID := uuid.New().String()

	_, err = h.DB.Exec(
		"INSERT INTO posts (id, user_id, title, content, image_path) VALUES (?, ?, ?, ?, ?)",
		postID, userID, title, content, imagePath,
	)
	if err != nil {
		h.InternalError(w, r)
		return
	}

	for _, catID := range categoryIDs {
		h.DB.Exec(
			"INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)",
			postID, catID,
		)
	}

	http.Redirect(w, r, "/post/"+postID, http.StatusSeeOther)
}
func (h *Handler) EditPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	userID, _ := utils.GetUserIDFromSession(r, h.DB)
	postID := r.FormValue("post_id")
	title := r.FormValue("title")
	content := r.FormValue("content")

	if title == "" || content == "" {
		http.Redirect(w, r, "/post/"+postID, http.StatusSeeOther)
		return
	}

	// Vérifie que c'est bien son post
	var ownerID string
	h.DB.QueryRow("SELECT user_id FROM posts WHERE id = ?", postID).Scan(&ownerID)
	if ownerID != userID {
		http.Error(w, "Non autorisé", http.StatusForbidden)
		return
	}

	h.DB.Exec("UPDATE posts SET title = ?, content = ? WHERE id = ?", title, content, postID)
	http.Redirect(w, r, "/post/"+postID, http.StatusSeeOther)
}

func (h *Handler) DeletePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	userID, _ := utils.GetUserIDFromSession(r, h.DB)
	postID := r.FormValue("post_id")

	var role string
	h.DB.QueryRow("SELECT role FROM users WHERE id = ?", userID).Scan(&role)

	if role != "admin" && role != "moderator" {
		http.Error(w, "Non autorisé", http.StatusForbidden)
		return
	}

	h.DB.Exec("DELETE FROM posts WHERE id = ?", postID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// --- Helpers ---

func getLikeCount(db *sql.DB, targetID, targetType string) (int, int) {
	var likes, dislikes int
	db.QueryRow("SELECT COUNT(*) FROM likes WHERE target_id = ? AND target_type = ? AND value = 1", targetID, targetType).Scan(&likes)
	db.QueryRow("SELECT COUNT(*) FROM likes WHERE target_id = ? AND target_type = ? AND value = -1", targetID, targetType).Scan(&dislikes)
	return likes, dislikes
}

func getUserVote(db *sql.DB, userID, targetID string) int {
	var value int
	db.QueryRow("SELECT value FROM likes WHERE user_id = ? AND target_id = ?", userID, targetID).Scan(&value)
	return value
}

func getPostCategories(db *sql.DB, postID string) []string {
	rows, err := db.Query(`
		SELECT c.name FROM categories c
		JOIN post_categories pc ON c.id = pc.category_id
		WHERE pc.post_id = ?`, postID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var cats []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		cats = append(cats, name)
	}
	return cats
}

func getAllCategories(db *sql.DB) []map[string]interface{} {
	rows, err := db.Query("SELECT id, name FROM categories")
	if err != nil {
		return nil
	}
	defer rows.Close()
	var cats []map[string]interface{}
	for rows.Next() {
		var id int
		var name string
		rows.Scan(&id, &name)
		cats = append(cats, map[string]interface{}{"ID": id, "Name": name})
	}
	return cats
}
