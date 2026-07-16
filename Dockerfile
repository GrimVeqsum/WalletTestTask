FROM golang:1.25-alpine AS source

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

FROM source AS tester

CMD ["go", "test", "./...", "-v", "-count=1"]

FROM source AS builder

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /server ./cmd/app

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /server /app/server

RUN adduser -D appuser
USER appuser

EXPOSE 8080

CMD ["/app/server"]