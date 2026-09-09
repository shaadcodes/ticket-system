# Ticket System API (Golang)

A simple backend service for a ticket system where users can register, login, create tickets, and update the status of their own tickets. Built with Go, containerized with Docker, and utilizing an in-memory data store.

## Live Deployment

- **Deployed URL:** `https://shaads-ticket-system-production.up.railway.app`
- **Health Check:** `https://shaads-ticket-system-production.up.railway.app/health`

## Local Run Instructions

To run this application locally using Docker:

1. Build the image:
   ```bash
   docker build -t ticket-system .
   ```
