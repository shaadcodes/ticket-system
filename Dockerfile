FROM golang:1.26-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o ticket-system .


FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/ticket-system .

EXPOSE 8080
CMD ["./ticket-system"]