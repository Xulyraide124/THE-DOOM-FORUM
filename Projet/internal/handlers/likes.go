package handlers

import (
	"forum/internal/utils"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

func (h *Handler) HandleLike(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	userID, _ := utils.GetUserIDFromSession(r, h.DB)
	targetID := r.FormValue("target_id")
	targetType := r.FormValue("target_type") // "post" ou "comment"
	valueStr := r.FormValue("value")         // "1" ou "-1"
	redirectTo := r.FormValue("redirect")

	value, err := strconv.Atoi(valueStr)
	if err != nil || (value != 1 && value != -1) {
		http.Error(w, "Valeur invalide", http.StatusBadRequest)
		return
	}

	if targetID == "" || (targetType != "post" && targetType != "comment") {
		http.Error(w, "Paramètres invalides", http.StatusBadRequest)
		return
	}

	// Vérifie si un vote existe déjà
	var existingID string
	var existingValue int
	err = h.DB.QueryRow(
		"SELECT id, value FROM likes WHERE user_id = ? AND target_id = ?",
		userID, targetID,
	).Scan(&existingID, &existingValue)

	if err == nil {
		// Vote existant
		if existingValue == value {
			// Même vote → on annule (toggle)
			h.DB.Exec("DELETE FROM likes WHERE id = ?", existingID)
		} else {
			// Vote différent → on change
			h.DB.Exec("UPDATE likes SET value = ? WHERE id = ?", value, existingID)
		}
	} else {
		// Nouveau vote
		h.DB.Exec(
			"INSERT INTO likes (id, user_id, target_id, target_type, value) VALUES (?, ?, ?, ?, ?)",
			uuid.New().String(), userID, targetID, targetType, value,
		)
	}

	if redirectTo == "" {
		redirectTo = "/"
	}
	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}
