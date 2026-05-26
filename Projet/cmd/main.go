package main

import (
	"fmt"
	"log"
	"net/http"

	"forum/internal/database"
	"forum/internal/handlers"
	"forum/internal/middleware"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load("../.env")
	// Init BDD
	db, err := database.InitDB("forum.db")
	if err != nil {
		log.Fatal("Erreur BDD:", err)
	}
	defer db.Close()

	// Migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatal("Erreur migrations:", err)
	}

	// Compte admin
	database.CreateAdminIfNotExists(db)

	// Handlers
	h := handlers.New(db)
	mux := http.NewServeMux()

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Routes publiques
	mux.HandleFunc("/", h.Index)
	mux.HandleFunc("/register", h.Register)
	mux.HandleFunc("/login", h.Login)
	mux.HandleFunc("/logout", h.Logout)
	mux.HandleFunc("/user/", h.ViewProfile)

	// Routes protégées
	mux.Handle("/profile/edit", middleware.Auth(db, http.HandlerFunc(h.EditProfile)))
	mux.Handle("/post/create", middleware.Auth(db, http.HandlerFunc(h.CreatePost)))
	mux.Handle("/post/delete", middleware.Auth(db, http.HandlerFunc(h.DeletePost)))
	mux.Handle("/post/edit", middleware.Auth(db, http.HandlerFunc(h.EditPost)))
	mux.Handle("/comment/create", middleware.Auth(db, http.HandlerFunc(h.CreateComment)))
	mux.Handle("/comment/delete", middleware.Auth(db, http.HandlerFunc(h.DeleteComment)))
	mux.Handle("/comment/edit", middleware.Auth(db, http.HandlerFunc(h.EditComment)))
	mux.Handle("/like", middleware.Auth(db, http.HandlerFunc(h.HandleLike)))

	// Routes admin/modération
	mux.Handle("/admin", middleware.Auth(db, http.HandlerFunc(h.AdminPanel)))
	mux.Handle("/admin/promote", middleware.Auth(db, http.HandlerFunc(h.PromoteUser)))
	mux.Handle("/admin/demote", middleware.Auth(db, http.HandlerFunc(h.DemoteUser)))
	mux.Handle("/admin/delete-user", middleware.Auth(db, http.HandlerFunc(h.DeleteUser)))
	mux.Handle("/admin/mute", middleware.Auth(db, http.HandlerFunc(h.MuteUser)))
	mux.Handle("/admin/unmute", middleware.Auth(db, http.HandlerFunc(h.UnmuteUser)))
	mux.Handle("/report", middleware.Auth(db, http.HandlerFunc(h.ReportContent)))
	mux.Handle("/moderate/delete", middleware.Auth(db, http.HandlerFunc(h.ModerateDelete)))

	// OAuth
	mux.HandleFunc("/auth/google/login", h.GoogleLogin)
	mux.HandleFunc("/auth/google/callback", h.GoogleCallback)

	// Route générique /post/ EN DERNIER
	mux.HandleFunc("/post/", h.ViewPost)
	fmt.Println("Serveur lancé sur http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
