package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/shaadcodes/ticket-system/config"
	"github.com/shaadcodes/ticket-system/models"
	"github.com/shaadcodes/ticket-system/store"
	"golang.org/x/crypto/bcrypt"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "\n\nInvalid request body!\n\n", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "\n\nUsername & Password are required!\n\n", http.StatusBadRequest)
		return
	}

	store.Store.Mu.RLock()
	for _, user := range store.Store.Users {
		if user.Username == req.Username {
			store.Store.Mu.RUnlock()
			http.Error(w, "\n\nUsername already exists!\n\n", http.StatusConflict)
			return
		}
	}
	store.Store.Mu.RUnlock()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "\n\nError hashing password!\n\n", http.StatusInternalServerError)
		return
	}

	userID := fmt.Sprintf("user_%d", time.Now().UnixNano())

	newUser := models.User{
		ID:           userID,
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
	}

	store.Store.Mu.Lock()
	store.Store.Users[userID] = newUser
	store.Store.Mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User created successfully!",
		"id":      userID,
	})

}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "\n\nInvalid request body!\n\n", http.StatusBadRequest)
		return
	}

	var foundUser models.User
	var userExists bool

	store.Store.Mu.RLock()
	for _, user := range store.Store.Users {
		if user.Username == req.Username {
			foundUser = user
			userExists = true
			break
		}
	}
	store.Store.Mu.RUnlock()

	if !userExists {
		http.Error(w, "\n\nUsername or password not found!\n\n", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(foundUser.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "\n\nUsername or password not found!\n\n", http.StatusUnauthorized)
		return
	}

	claims := jwt.MapClaims{
		"user_id": foundUser.ID,
		"expires": time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(config.JWTKey)

	if err != nil {
		http.Error(w, "\n\nError signing token!\n\n", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
	})
}
