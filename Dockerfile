FROM golang:1.24.0 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main /app/cmd/app/main.go

FROM debian:bookworm-slim
WORKDIR /
RUN apt-get update && apt-get install -y curl ca-certificates && \
    rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/main ./main
COPY --from=builder /app/migrations ./migrations
CMD [ "./main" ]