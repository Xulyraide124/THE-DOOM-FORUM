package handlers

import (
	"forum/internal/utils"
	"net/http"

	"github.com/google/uuid"
)

func (h *Handler) getUserRole(r *http.Request) string {
	userID, err := utils.GetUserIDFromSession(r, h.DB)
	if err != nil {
		return "guest"
	}
	var role string
	h.DB.QueryRow("SELECT role FROM users WHERE id = ?", userID).Scan(&role)
	return role
}

func (h *Handler) AdminPanel(w http.ResponseWriter, r *http.Request) {
	role := h.getUserRole(r)
	if role != "admin" {
		http.Error(w, "Accès interdit", http.StatusForbidden)
		return
	}

	rows, err := h.DB.Query("SELECT id, username, email, role, muted, created_at FROM users ORDER BY created_at DESC")
	if err != nil {
		h.InternalError(w, r)
		return
	}
	defer rows.Close()

	type UserRow struct {
		ID        string
		Username  string
		Email     string
		Role      string
		Muted     int
		CreatedAt string
	}

	var users []UserRow
	for rows.Next() {
		var u UserRow
		var createdAt string
		rows.Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.Muted, &createdAt)
		if len(createdAt) >= 10 {
			u.CreatedAt = createdAt[:10]
		} else {
			u.CreatedAt = createdAt
		}
		users = append(users, u)
	}

	reportRows, err := h.DB.Query(`
		SELECT r.id, u.username, r.target_id, r.target_type, r.reason, r.created_at
		FROM reports r JOIN users u ON r.reporter_id = u.id
		ORDER BY r.created_at DESC
	`)
	if err != nil {
		h.InternalError(w, r)
		return
	}
	defer reportRows.Close()

	type ReportRow struct {
		ID         string
		Reporter   string
		TargetID   string
		TargetType string
		Reason     string
		CreatedAt  string
	}

	var reports []ReportRow
	for reportRows.Next() {
		var rep ReportRow
		var createdAt string
		reportRows.Scan(&rep.ID, &rep.Reporter, &rep.TargetID, &rep.TargetType, &rep.Reason, &createdAt)
		if len(createdAt) >= 10 {
			rep.CreatedAt = createdAt[:10]
		} else {
			rep.CreatedAt = createdAt
		}
		reports = append(reports, rep)
	}

	userID, _ := utils.GetUserIDFromSession(r, h.DB)
	var username string
	h.DB.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)

	h.Templates["admin"].ExecuteTemplate(w, "base", map[string]interface{}{
		"Users":    users,
		"Reports":  reports,
		"UserID":   userID,
		"Username": username,
		"Role":     role,
	})
}

func (h *Handler) PromoteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}
	role := h.getUserRole(r)
	if role != "admin" {
		http.Error(w, "Accès interdit", http.StatusForbidden)
		return
	}
	targetID := r.FormValue("user_id")
	h.DB.Exec("UPDATE users SET role = 'moderator' WHERE id = ? AND role = 'user'", targetID)
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *Handler) DemoteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}
	role := h.getUserRole(r)
	if role != "admin" {
		http.Error(w, "Accès interdit", http.StatusForbidden)
		return
	}
	targetID := r.FormValue("user_id")
	h.DB.Exec("UPDATE users SET role = 'user' WHERE id = ? AND role = 'moderator'", targetID)
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *Handler) ReportContent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}
	userID, _ := utils.GetUserIDFromSession(r, h.DB)
	targetID := r.FormValue("target_id")
	targetType := r.FormValue("target_type")
	reason := r.FormValue("reason")
	redirectTo := r.FormValue("redirect")

	if targetID == "" || reason == "" {
		if redirectTo == "" {
			redirectTo = "/"
		}
		http.Redirect(w, r, redirectTo, http.StatusSeeOther)
		return
	}

	h.DB.Exec(
		"INSERT INTO reports (id, reporter_id, target_id, target_type, reason) VALUES (?, ?, ?, ?, ?)",
		uuid.New().String(), userID, targetID, targetType, reason,
	)

	if redirectTo == "" {
		redirectTo = "/"
	}
	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}

func (h *Handler) ModerateDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}
	role := h.getUserRole(r)
	if role != "admin" && role != "moderator" {
		http.Error(w, "Accès interdit", http.StatusForbidden)
		return
	}
	targetID := r.FormValue("target_id")
	targetType := r.FormValue("target_type")
	redirectTo := r.FormValue("redirect")

	if targetType == "post" {
		h.DB.Exec("DELETE FROM posts WHERE id = ?", targetID)
	} else if targetType == "comment" {
		h.DB.Exec("DELETE FROM comments WHERE id = ?", targetID)
	}

	if redirectTo == "" {
		redirectTo = "/"
	}
	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}
	role := h.getUserRole(r)
	if role != "admin" {
		http.Error(w, "Accès interdit", http.StatusForbidden)
		return
	}
	targetID := r.FormValue("user_id")

	var targetRole string
	h.DB.QueryRow("SELECT role FROM users WHERE id = ?", targetID).Scan(&targetRole)
	if targetRole == "admin" {
		http.Error(w, "Impossible de supprimer un admin", http.StatusForbidden)
		return
	}

	h.DB.Exec("DELETE FROM users WHERE id = ?", targetID)
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *Handler) MuteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}
	role := h.getUserRole(r)
	if role != "admin" && role != "moderator" {
		http.Error(w, "Accès interdit", http.StatusForbidden)
		return
	}
	targetID := r.FormValue("user_id")

	var targetRole string
	h.DB.QueryRow("SELECT role FROM users WHERE id = ?", targetID).Scan(&targetRole)
	if targetRole == "admin" || (role == "moderator" && targetRole == "moderator") {
		http.Error(w, "Action non autorisée", http.StatusForbidden)
		return
	}

	h.DB.Exec("UPDATE users SET muted = 1 WHERE id = ?", targetID)
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *Handler) UnmuteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}
	role := h.getUserRole(r)
	if role != "admin" && role != "moderator" {
		http.Error(w, "Accès interdit", http.StatusForbidden)
		return
	}
	targetID := r.FormValue("user_id")
	h.DB.Exec("UPDATE users SET muted = 0 WHERE id = ?", targetID)
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
