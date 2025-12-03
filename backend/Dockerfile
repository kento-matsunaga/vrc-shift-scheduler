FROM golang:1.22

WORKDIR /app

# Copy all source code
COPY . .

# Download dependencies and generate go.sum
RUN go mod tidy

# Default command (can be overridden in docker-compose)
CMD ["go", "run", "./cmd/api"]

