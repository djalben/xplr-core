# Stage 1: Build the Go application
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Copy go module files from project root
COPY go.mod go.sum ./

RUN go mod download

# Copy backend source and api/ (needed for module completeness)
COPY backend/ ./backend/
COPY api/ ./api/

# Build the server binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./backend/cmd/server

# Stage 2: Minimal runtime image
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]
