package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/shaadcodes/ticket-system/config"
	"github.com/shaadcodes/ticket-system/models"
	"github.com/shaadcodes/ticket-system/store"
)

func CreateTicketHandler(w http.ResponseWriter, r *http.Request) {

	userID, ok := r.Context().Value(config.UserIDKey).(string)
	if !ok {
		http.Error(w, "\n\nUser identity missing!\n\n", http.StatusInternalServerError)
		return
	}

	var req models.CreateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "\n\nInvalid request body!\n\n", http.StatusBadRequest)
		return
	}

	ticketID := fmt.Sprintf("tkt_%d", time.Now().UnixNano())
	newTicket := models.Ticket{
		ID:          ticketID,
		Title:       req.Title,
		Description: req.Description,
		Status:      "open",
		UserID:      userID,
	}

	store.Store.Mu.Lock()
	store.Store.Tickets[ticketID] = newTicket
	store.Store.Mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTicket)

}

func GetTicketHandler(w http.ResponseWriter, r *http.Request) {

	userID, ok := r.Context().Value(config.UserIDKey).(string)
	if !ok {
		http.Error(w, "\n\nUser ID missing!\n\n", http.StatusInternalServerError)
		return
	}

	userTickets := make([]models.Ticket, 0)

	store.Store.Mu.RLock()
	for _, ticket := range store.Store.Tickets {
		if ticket.UserID == userID {
			userTickets = append(userTickets, ticket)
		}
	}
	store.Store.Mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userTickets)
}

func GetTicketByIDHandler(w http.ResponseWriter, r *http.Request) {

	userID, ok := r.Context().Value(config.UserIDKey).(string)
	if !ok {
		http.Error(w, "\n\nUser ID missing!\n\n", http.StatusInternalServerError)
		return
	}

	ticketID := r.PathValue("id")

	store.Store.Mu.RLock()
	ticket, exists := store.Store.Tickets[ticketID]
	store.Store.Mu.RUnlock()

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

func UpdateTicketHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(config.UserIDKey).(string)
	if !ok {
		http.Error(w, "\n\nUser identity missing!\n\n", http.StatusInternalServerError)
	}

	ticketID := r.PathValue("id")

	var req models.UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "\n\nInvalid request body!\n\n", http.StatusBadRequest)
		return
	}

	validateStatuses := map[string]bool{"open": true, "in_progress": true, "closed": true}
	if !validateStatuses[req.Status] {
		http.Error(w, "\n\nInvalid status! Must be 'open', 'in_progress' or 'closed'", http.StatusBadRequest)
		return
	}

	store.Store.Mu.Lock()
	defer store.Store.Mu.Unlock()

	ticket, exists := store.Store.Tickets[ticketID]
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
	store.Store.Tickets[ticketID] = ticket

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ticket)

}
