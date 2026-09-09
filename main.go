package main

import (
	"log"
	"net/http"

	"github.com/shaadcodes/ticket-system/handlers"
	"github.com/shaadcodes/ticket-system/middleware"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handlers.HealthHandler)
	mux.HandleFunc("POST /auth/register", handlers.RegisterHandler)
	mux.HandleFunc("POST /auth/login", handlers.LoginHandler)
	mux.HandleFunc("POST /tickets", middleware.AuthMiddleware(handlers.CreateTicketHandler))
	mux.HandleFunc("GET /tickets", middleware.AuthMiddleware(handlers.GetTicketHandler))
	mux.HandleFunc("GET /tickets/{id}", middleware.AuthMiddleware(handlers.GetTicketByIDHandler))
	mux.HandleFunc("PATCH /tickets/{id}/status", middleware.AuthMiddleware(handlers.UpdateTicketHandler))

	log.Println("Starting server on port: 8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Unable to start server: %v", err)
	}
}
