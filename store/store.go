package store

import (
	"github.com/shaadcodes/ticket-system/models"
)

var Store = models.InMemoryStore{
	Users:   make(map[string]models.User),
	Tickets: make(map[string]models.Ticket),
}
