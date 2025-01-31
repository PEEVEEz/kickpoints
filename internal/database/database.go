package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
)

// Service represents a service that interacts with a database.
type Service interface {
	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	Close() error

	// AddPoints adds points to a user and returns the new amount.
	AddPoints(username string, points int) (int, error)

	// RemovePoints removes points from a user and returns success and the new amount.
	RemovePoints(username string, points int) (bool, int, error)

	// GetPoints returns the points of a user.
	GetPoints(username string) (int, error)

	// Get all users with points
	GetAllUsers() ([]User, error)
}

type service struct {
	db *sql.DB
}

var (
	database   = os.Getenv("DATABASE_DATABASE")
	password   = os.Getenv("DATABASE_PASSWORD")
	username   = os.Getenv("DATABASE_USERNAME")
	port       = os.Getenv("DATABASE_PORT")
	host       = os.Getenv("DATABASE_HOST")
	schema     = os.Getenv("DATABASE_SCHEMA")
	dbInstance *service
)

func New() Service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s", username, password, host, port, database, schema)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}

	// Run migrations
	if err := migrate(db); err != nil {
		log.Fatal(err)
	}

	dbInstance = &service{
		db: db,
	}

	return dbInstance
}

// User represents the user table in the database.
type User struct {
	ID        uint      `db:"id" json:"id"`
	Username  string    `db:"username" json:"username"`
	Points    int       `db:"points" json:"points"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

// migrate creates the users table if it doesn't exist.
func migrate(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) NOT NULL,
			points INT DEFAULT 0,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT unique_username UNIQUE (username)
		);
	`

	_, err := db.Exec(query)
	return err
}

// GetAllUsers returns all users with points
func (s *service) GetAllUsers() ([]User, error) {
	query := `SELECT id, username, points, updated_at, created_at FROM users`

	rows, err := s.db.Query(query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []User

	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Username, &user.Points, &user.UpdatedAt, &user.CreatedAt); err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}

// AddPoints adds points to a user and returns the new amount.
func (s *service) AddPoints(username string, points int) (int, error) {
	query := `
		INSERT INTO users (username, points, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
		ON CONFLICT (username) 
		DO UPDATE 
		SET points = users.points + EXCLUDED.points, updated_at = CURRENT_TIMESTAMP
		RETURNING points
	`
	var newAmount int
	err := s.db.QueryRow(query, username, points).Scan(&newAmount)

	return newAmount, err
}

// RemovePoints removes points from a user and returns success and the new amount.
func (s *service) RemovePoints(username string, points int) (bool, int, error) {
	query := `
		UPDATE users
		SET points = points - $1, updated_at = CURRENT_TIMESTAMP
		WHERE username = $2 AND channel = $3 AND points >= $1
		RETURNING points
    `

	var newAmount int
	err := s.db.QueryRow(query, points, username).Scan(&newAmount)
	if err == sql.ErrNoRows {
		return false, 0, nil
	}

	return err == nil, newAmount, err
}

// GetPoints returns the points of a user.
func (s *service) GetPoints(username string) (int, error) {
	query := `
		SELECT points
		FROM users
		WHERE channel = $1 AND username = $2
    `

	var points int
	err := s.db.QueryRow(query, username).Scan(&points)

	return points, err
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *service) Close() error {
	log.Printf("Disconnected from database: %s", database)
	return s.db.Close()
}
