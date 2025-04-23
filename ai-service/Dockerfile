FROM golang:1.23.4-alpine3.21 AS builder
WORKDIR /app
COPY . .

RUN go build -o bin/ai main.go


FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/bin/ai .
EXPOSE 8080
CMD ["./ai", "start", "--port", "8080", "--debug", "--swagger"]
