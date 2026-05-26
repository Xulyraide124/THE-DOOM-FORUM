package middleware

import (
	"database/sql"
	"forum/internal/utils"
	"net/http"
)

func Auth(db *sql.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := utils.GetUserIDFromSession(r, db)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
