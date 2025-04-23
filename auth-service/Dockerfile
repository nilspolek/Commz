FROM golang:1.23.4-alpine3.21 AS builder
WORKDIR /app
COPY . .
# RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN go build -o bin/auth main.go


FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/bin/auth .
EXPOSE 8080
CMD ["./auth", "start", "--port", "8080", "--debug", "--swagger"]

