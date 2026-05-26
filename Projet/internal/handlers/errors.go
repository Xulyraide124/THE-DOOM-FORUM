package handlers

import "net/http"

func (h *Handler) NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	h.Templates["error"].ExecuteTemplate(w, "base", map[string]interface{}{
		"Code":    404,
		"Message": "Page introuvable",
	})
}

func (h *Handler) InternalError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	h.Templates["error"].ExecuteTemplate(w, "base", map[string]interface{}{
		"Code":    500,
		"Message": "Erreur interne du serveur",
	})
}
