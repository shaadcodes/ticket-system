package config

type contextKey string

var JWTKey = []byte("secret_assignment_key")

const UserIDKey contextKey = "userID"
