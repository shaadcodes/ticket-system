package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"` //IGNORE SENDING PASSWORD HASHES OVER API RESPONSE
}

type Ticket struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	UserID      string `json:"user_id"`
}

type InMemoryStore struct {
	mu      sync.RWMutex
	Users   map[string]User
	Tickets map[string]Ticket
}

var store = InMemoryStore{
	Users:   make(map[string]User),
	Tickets: make(map[string]Ticket),
}

type HealthResponse struct {
	Status string `json:"status"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := HealthResponse{Status: "ok"}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "\n\nFailed to encode response!\n\n", http.StatusInternalServerError)
	}
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "\n\nInvalid request body!\n\n", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "\n\nUsername & Password are required!\n\n", http.StatusBadRequest)
		return
	}

	store.mu.RLock()
	for _, user := range store.Users {
		if user.Username == req.Username {
			store.mu.RUnlock()
			http.Error(w, "\n\nUsername already exists!\n\n", http.StatusConflict)
			return
		}
	}
	store.mu.RUnlock()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "\n\nError hashing password!\n\n", http.StatusInternalServerError)
		return
	}

	userID := fmt.Sprintf("user_%d", time.Now().UnixNano())

	newUser := User{
		ID:           userID,
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
	}

	store.mu.Lock()
	store.Users[userID] = newUser
	store.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User created successfully!",
		"id":      userID,
	})

}

var jwtKey = []byte("secret_assignment_key")

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "\n\nInvalid request body!\n\n", http.StatusBadRequest)
		return
	}

	var foundUser User
	var userExists bool

	store.mu.RLock()
	for _, user := range store.Users {
		if user.Username == req.Username {
			foundUser = user
			userExists = true
			break
		}
	}
	store.mu.RUnlock()

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
	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		http.Error(w, "\n\nError signing token!\n\n", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
	})
}

type contextKey string

const userIDKey contextKey = "userID"

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "\n\nAuthorization header required!\n\n", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			http.Error(w, "\n\nInvalid token format!\n\n", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Signing method does not match!\n\n")
			}
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "\n\nInvalid token!\n\n", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "\n\nInvalid token claims!\n\n", http.StatusUnauthorized)
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok {
			http.Error(w, "\n\nNo user ID in token!\n\n", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

type CreateTicketRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func createTicketHandler(w http.ResponseWriter, r *http.Request) {

	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok {
		http.Error(w, "\n\nUser identity missing!\n\n", http.StatusInternalServerError)
		return
	}

	var req CreateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "\n\nInvalid request body!\n\n", http.StatusBadRequest)
		return
	}

	ticketID := fmt.Sprintf("tkt_%d", time.Now().UnixNano())
	newTicket := Ticket{
		ID:          ticketID,
		Title:       req.Title,
		Description: req.Description,
		Status:      "open",
		UserID:      userID,
	}

	store.mu.Lock()
	store.Tickets[ticketID] = newTicket
	store.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTicket)

}

func getTicketHandler(w http.ResponseWriter, r *http.Request) {

	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok {
		http.Error(w, "\n\nUser ID missing!\n\n", http.StatusInternalServerError)
		return
	}

	userTickets := make([]Ticket, 0)

	store.mu.RLock()
	for _, ticket := range store.Tickets {
		if ticket.UserID == userID {
			userTickets = append(userTickets, ticket)
		}
	}
	store.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userTickets)
}

func getTicketByIDHandler(w http.ResponseWriter, r *http.Request) {

	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok {
		http.Error(w, "\n\nUser ID missing!\n\n", http.StatusInternalServerError)
		return
	}

	ticketID := r.PathValue("id")

	store.mu.RLock()
	ticket, exists := store.Tickets[ticketID]
	store.mu.RUnlock()

	if !exists {
		http.Error(w, "\n\nTicket not found!\n\n", http.StatusNotFound)
		return
	}

	if ticket.UserID != userID {
		http.Error(w, "\n\nThis ticket is not yours!\n\n", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ticket)

}

type UpdateRequest struct {
	Status string `json:"status"`
}

func updateTicketHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok {
		http.Error(w, "\n\nUser identity missing!\n\n", http.StatusInternalServerError)
	}

	ticketID := r.PathValue("id")

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "\n\nInvalid request body!\n\n", http.StatusBadRequest)
		return
	}

	validateStatuses := map[string]bool{"open": true, "in_progress": true, "closed": true}
	if !validateStatuses[req.Status] {
		http.Error(w, "\n\nInvalid status! Must be 'open', 'in_progress' or 'closed'", http.StatusBadRequest)
		return
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	ticket, exists := store.Tickets[ticketID]
	if !exists {
		http.Error(w, "\n\nTicket missing!\n\n", http.StatusNotFound)
		return
	}

	if ticket.UserID != userID {
		http.Error(w, "\n\nThis ticket is not yours!\n\n", http.StatusForbidden)
		return
	}

	if ticket.Status == "closed" {
		http.Error(w, "\n\nClosed tickets cannot be updated!\n\n", http.StatusBadRequest)
		return
	}

	if ticket.Status == "open" && req.Status == "closed" {
		http.Error(w, "\n\nTicket must be in 'in_progress' before closing!\n\n", http.StatusBadRequest)
		return
	}

	ticket.Status = req.Status
	store.Tickets[ticketID] = ticket

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ticket)

}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("POST /auth/register", registerHandler)
	mux.HandleFunc("POST /auth/login", loginHandler)
	mux.HandleFunc("POST /tickets", authMiddleware(createTicketHandler))
	mux.HandleFunc("GET /tickets", authMiddleware(getTicketHandler))
	mux.HandleFunc("GET /tickets/{id}", authMiddleware(getTicketByIDHandler))
	mux.HandleFunc("PATCH /tickets/{id}/status", authMiddleware(updateTicketHandler))

	log.Println("Starting server on port: 8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Unable to start server: %v", err)
	}
}
