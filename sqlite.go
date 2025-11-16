package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func initDB(dbPath string) error {
	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	// Enable WAL mode for better concurrency
	_, err = db.Exec("PRAGMA journal_mode=WAL")
	if err != nil {
		return err
	}

	// Create tables
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS lists (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id),
			UNIQUE(user_id, name)
		);

		CREATE TABLE IF NOT EXISTS photos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			list_id INTEGER NOT NULL,
			filename TEXT NOT NULL,
			original_filename TEXT NOT NULL,
			date TIMESTAMP,
			location TEXT,
			caption TEXT,
			width INTEGER,
			height INTEGER,
			file_size INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (list_id) REFERENCES lists(id)
		);

		CREATE INDEX IF NOT EXISTS idx_photos_list_id ON photos(list_id);
		CREATE INDEX IF NOT EXISTS idx_lists_user_id ON lists(user_id);
	`)
	return err
}

func getOrCreateUser(username string) (int64, error) {
	var userID int64
	err := db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err == sql.ErrNoRows {
		result, err := db.Exec("INSERT INTO users (username) VALUES (?)", username)
		if err != nil {
			return 0, err
		}
		return result.LastInsertId()
	}
	return userID, err
}

func getOrCreateList(userID int64, listName string) (int64, error) {
	var listID int64
	err := db.QueryRow("SELECT id FROM lists WHERE user_id = ? AND name = ?", userID, listName).Scan(&listID)
	if err == sql.ErrNoRows {
		result, err := db.Exec("INSERT INTO lists (user_id, name) VALUES (?, ?)", userID, listName)
		if err != nil {
			return 0, err
		}
		return result.LastInsertId()
	}
	return listID, err
}

func savePhotoToDB(listID int64, photo Photo, width, height, fileSize int) error {
	_, err := db.Exec(`
		INSERT INTO photos (list_id, filename, original_filename, date, location, caption, width, height, file_size)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, listID, photo.Filename, photo.Filename, photo.Date, photo.Location, "", width, height, fileSize)
	return err
}

func getListPhotos(username, listName string) ([]Photo, error) {
	rows, err := db.Query(`
		SELECT p.filename, p.date, p.location
		FROM photos p
		JOIN lists l ON p.list_id = l.id
		JOIN users u ON l.user_id = u.id
		WHERE u.username = ? AND l.name = ?
		ORDER BY p.date DESC
	`, username, listName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var photos []Photo
	for rows.Next() {
		var p Photo
		if err := rows.Scan(&p.Filename, &p.Date, &p.Location); err != nil {
			return nil, err
		}
		photos = append(photos, p)
	}
	return photos, rows.Err()
}

func getUserLists(username string) ([]string, error) {
	rows, err := db.Query(`
		SELECT l.name
		FROM lists l
		JOIN users u ON l.user_id = u.id
		WHERE u.username = ?
		ORDER BY l.created_at DESC
	`, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lists []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		lists = append(lists, name)
	}
	return lists, rows.Err()
}

func closeDB() error {
	if db != nil {
		return db.Close()
	}
	return nil
}
