# --- Build ---
FROM golang:1.25-alpine AS builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY backend/ .
RUN go mod download
ARG VERSION=dev
RUN CGO_ENABLED=1 go build -ldflags "-X main.Version=${VERSION}" -o roadlog .

# --- Runtime ---
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/roadlog .
RUN mkdir -p /data
VOLUME /data
EXPOSE 3000
ENV DATA_DIR=/data
LABEL net.unraid.docker.icon="https://raw.githubusercontent.com/stomko11/roadlog/main/icon.png"
LABEL net.unraid.docker.webui="http://[IP]:[PORT:3000]/"
CMD ["./roadlog"]
