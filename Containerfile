FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/alerts-adapter ./cmd/alerts-adapter

FROM golang:1.26

COPY --from=builder /bin/alerts-adapter /usr/local/bin/alerts-adapter

USER 1001

ENTRYPOINT ["alerts-adapter"]
