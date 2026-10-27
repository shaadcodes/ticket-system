# Ticket System API (Golang)

A simple backend service for a ticket system where users can register, login, create tickets, and update the status of their own tickets. Built with Go, containerized with Docker, and utilizing an in-memory data store.

## Live Deployment

- **Deployed URL:** `https://ticket-system-api-6qha.onrender.com`
- **Health Check:** `https://ticket-system-api-6qha.onrender.com/health`

## Local Run Instructions

To run this application locally using Docker:

1. Build the image:
   ```bash
   docker build -t ticket-system .
   ```

## Screenshots

### Health check: ok
![Health check: ok](./screenshots/Health%20Check.png)

### Register new user
![Register new user](./screenshots/Register%20new%20user.png)

### Register new user: validation
![Register new user: validation](./screenshots/User%20validation.png)

### Login and JWT
![Login and JWT](./screenshots/Login%20and%20jwt.png)

### Raising tickets
![Raising tickets](./screenshots/Raise%20a%20ticket.png)

### Retrieving raised tickets
![Retrieving raised tickets](./screenshots/Retrieveing%20raised%20tickets.png)

### Updating ticket status: following strict progress
![Updating ticket status: following strict progress](./screenshots/Update%20ticket%20status(strict%20flow).png)

### Updating ticket status: updating to in_progress
![Updating ticket status: updating to in_progress](./screenshots/Update%20ticket%20in_progress.png)

### Updating ticket status: updating to closed
![Updating ticket status: updating to closed](./screenshots/Update%20ticket%20closed.png)

### Updating ticket status: closed tickets cannot be updated
![Updating ticket status: closed tickets cannot be updated](./screenshots/Closed%20tickets%20immune%20to%20updates.png)
