FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG SERVICE_PATH

RUN go build -o /main ./cmd/${SERVICE_PATH}

FROM alpine:3.18

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /main .

# COPY config/ ./config/

EXPOSE 8080 50051 50052

CMD ["./main"]