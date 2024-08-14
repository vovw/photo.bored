package user
import (
	"encoding/json"
	"net/http"
	"github.com/google/uuid"
)
 type User struct{
	ID          uuid.UUID   `json:"id"`
	Username   string      `json:"first_name"`
    Email       string      `json:"email"`
    Password    string    `json:"password"`  
 }
 func (m *Model) RegisterUser(w http.ResponseWriter, r *http.Request) {
	// Define a struct to hold the incoming JSON data
	var input struct {
		Username string `json:"Username"`
		Email    string `json:"Email"`
		Password string `json:"Password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	if input.Username == "" || input.Email == "" || input.Password == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}
	if err := CheckPasswordLevel(input.Password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := m.userstore.CheckDBConnection(); err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	hashedPassword, err := PasswordHash(input.Password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}
	user := &User{
		Username: input.Username,
		Email:    input.Email,
		Password: hashedPassword,
	}
	if err := m.userstore.CreateUser(user); err != nil {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User registered successfully"))
}
