FROM golang:1.26 as dev
WORKDIR /app

RUN go install github.com/air-verse/air@v1.61.7

COPY go.mod go.sum ./
RUN go mod download
COPY . .

CMD ["air", "-c", ".air.toml"]

FROM golang:1.26 as builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o migrate ./cmd/migrate
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/migrate ./migrate
COPY --from=builder /app/main ./main
COPY ./migrations /root/migrations

EXPOSE 8080
ENTRYPOINT ["sh", "-c", "./main"]
