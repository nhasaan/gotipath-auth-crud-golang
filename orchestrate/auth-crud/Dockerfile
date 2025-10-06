# Build Golang application Image
FROM golang:1.23.3 AS builder

WORKDIR /app

COPY source/auth-crud/go.mod source/auth-crud/go.sum ./
RUN go mod download && go mod verify

COPY source/auth-crud/ .

RUN go build -o main .

# Run Golang application Image
FROM golang:1.23.3 AS runner

WORKDIR /app

COPY --from=builder /app/main .

CMD ["./main"]