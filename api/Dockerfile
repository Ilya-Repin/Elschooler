FROM golang:1.23.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod tidy

COPY . .

RUN go build -o /app/cmd/server/server cmd/server/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/cmd/server/server /app/server

EXPOSE 9090
EXPOSE 44044

CMD ["/app/server"]
