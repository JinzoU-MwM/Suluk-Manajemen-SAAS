FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG SERVICE
RUN CGO_ENABLED=0 go build -o /app/service ./cmd/${SERVICE}

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/service .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080 50051 50052 50053 50054 50055 50056 50057 50058

ENTRYPOINT ["./service"]
