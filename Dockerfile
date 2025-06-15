# Stage 1: Build binary
FROM golang:1.23.4-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o trade-engine .

# Stage 2: Minimal runtime image
FROM alpine:latest

RUN apk add --no-cache bash ca-certificates

WORKDIR /app

COPY --from=builder /app/trade-engine .
COPY ./database /app/database
COPY ./exchange.yaml /app/exchange.yaml
COPY .env.docker /app/.env

EXPOSE 9000

ENTRYPOINT ["./trade-engine"]
CMD ["engine"]
