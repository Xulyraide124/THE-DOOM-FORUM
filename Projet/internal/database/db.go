package database

import (
	"database/sql"
	"log"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

func InitDB(filepath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return nil, err
	}

	return db, nil
}

func CreateAdminIfNotExists(db *sql.DB) {
	var count int
	db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&count)
	if count > 0 {
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte("admin1234"), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Erreur création admin:", err)
		return
	}

	_, err = db.Exec(
		"INSERT INTO users (id, email, username, password, role) VALUES (?, ?, ?, ?, ?)",
		uuid.New().String(), "admin@doom.fr", "admin", string(hashed), "admin",
	)
	if err != nil {
		log.Println("Erreur insertion admin:", err)
		return
	}
	log.Println("Compte admin créé — email: admin@doom.fr / password: admin1234")
}
