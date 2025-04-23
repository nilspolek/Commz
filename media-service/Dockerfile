FROM golang:1.23.4-alpine3.21 AS builder
WORKDIR /app
COPY . .
# RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN go build -o bin/media main.go


FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/bin/media .
EXPOSE 8080
CMD ["./media", "start", "--port", "8080"]

