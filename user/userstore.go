package user
import (
	"database/sql"
	"log"
  "fmt"
)
type UserStore struct {
	db *sql.DB
}
func (us *UserStore) CreateUser(user *User) error {
    query := "INSERT INTO users (Username, Email, Password, credit, area, address) VALUES ($1, $2, $3"
    _, err := us.db.Exec(query, user.Username, user.Email, user.Password)
    if err != nil {
		log.Printf("Error creating user: %v", err)
         // Log the error
    }
    return err
}
func (s *UserStore) CheckDBConnection() error {
	err := s.db.Ping()
	if err != nil {
		return fmt.Errorf("failed to connect to the database: %v", err)
	}
	return nil
}