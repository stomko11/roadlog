# --- Build ---
FROM golang:1.25-alpine AS builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY backend/ .
RUN go mod download
RUN CGO_ENABLED=1 go build -o roadlog .

# --- Runtime ---
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/roadlog .
RUN mkdir -p /data
EXPOSE 3000
ENV DATA_DIR=/data
CMD ["./roadlog"]
