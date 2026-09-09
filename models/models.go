package models

import (
	"sync"
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
	Mu      sync.RWMutex
	Users   map[string]User
	Tickets map[string]Ticket
}

type HealthResponse struct {
	Status string `json:"status"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CreateTicketRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type UpdateRequest struct {
	Status string `json:"status"`
}
