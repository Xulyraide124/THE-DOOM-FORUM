package database

import "database/sql"

func RunMigrations(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'user',
			muted INTEGER NOT NULL DEFAULT 0,
			avatar TEXT DEFAULT '',
			bio TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			expires_at DATETIME NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS posts (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			image_path TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS post_categories (
			post_id TEXT NOT NULL,
			category_id INTEGER NOT NULL,
			PRIMARY KEY (post_id, category_id),
			FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
			FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS comments (
			id TEXT PRIMARY KEY,
			post_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			content TEXT NOT NULL,
			image_path TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS likes (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			target_id TEXT NOT NULL,
			target_type TEXT NOT NULL CHECK(target_type IN ('post','comment')),
			value INTEGER NOT NULL CHECK(value IN (1,-1)),
			UNIQUE(user_id, target_id),
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS reports (
			id TEXT PRIMARY KEY,
			reporter_id TEXT NOT NULL,
			target_id TEXT NOT NULL,
			target_type TEXT NOT NULL CHECK(target_type IN ('post','comment')),
			reason TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (reporter_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`INSERT OR IGNORE INTO categories (name) VALUES 
			('Général'), ('Tech'), ('Gaming'), ('Musique'), ('Cinéma')`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}
