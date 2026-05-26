package handlers

import (
	"forum/internal/models"
	"forum/internal/utils"
	"net/http"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.Templates["register"].ExecuteTemplate(w, "base", nil)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	email := r.FormValue("email")
	username := r.FormValue("username")
	password := r.FormValue("password")

	if email == "" || username == "" || password == "" {
		h.Templates["register"].ExecuteTemplate(w, "base", map[string]interface{}{
			"Error": "Tous les champs sont obligatoires",
		})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		h.InternalError(w, r)
		return
	}

	user := models.User{
		ID:       uuid.New().String(),
		Email:    email,
		Username: username,
		Password: string(hashed),
	}

	_, err = h.DB.Exec(
		"INSERT INTO users (id, email, username, password) VALUES (?, ?, ?, ?)",
		user.ID, user.Email, user.Username, user.Password,
	)
	if err != nil {
		h.Templates["register"].ExecuteTemplate(w, "base", map[string]interface{}{
			"Error": "Email ou nom d'utilisateur déjà pris",
		})
		return
	}

	if err := utils.CreateSession(w, h.DB, user.ID); err != nil {
		h.InternalError(w, r)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.Templates["login"].ExecuteTemplate(w, "base", nil)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	if email == "" || password == "" {
		h.Templates["login"].ExecuteTemplate(w, "base", map[string]interface{}{
			"Error": "Tous les champs sont obligatoires",
		})
		return
	}

	var user models.User
	err := h.DB.QueryRow(
		"SELECT id, password FROM users WHERE email = ?", email,
	).Scan(&user.ID, &user.Password)
	if err != nil {
		h.Templates["login"].ExecuteTemplate(w, "base", map[string]interface{}{
			"Error": "Email ou mot de passe incorrect",
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		h.Templates["login"].ExecuteTemplate(w, "base", map[string]interface{}{
			"Error": "Email ou mot de passe incorrect",
		})
		return
	}

	if err := utils.CreateSession(w, h.DB, user.ID); err != nil {
		h.InternalError(w, r)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	utils.DestroySession(w, h.DB, r)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
