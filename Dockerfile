FROM golang:1.24.0-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build /app/cmd/app/main.go

FROM alpine
WORKDIR /app
COPY --from=builder /app/main ./main
COPY --from=builder /app/migrations ./migrations
CMD [ "./main" ]