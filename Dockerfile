FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o antibruteforce ./cmd/server

FROM scratch
COPY --from=builder /app/antibruteforce /antibruteforce
ENTRYPOINT ["/antibruteforce"]
