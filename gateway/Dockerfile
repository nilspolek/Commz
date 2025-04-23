FROM golang:1.23.4-alpine3.21 AS builder
WORKDIR /app
COPY . .
# RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN go build -o bin/gateway main.go


FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/bin/gateway .
EXPOSE 8080
CMD ["./gateway", "start", "--debug", "--port", "8080"]

