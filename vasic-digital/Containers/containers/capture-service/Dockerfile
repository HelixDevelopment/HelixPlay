# vasic-digital/Containers/containers/capture-service/Dockerfile
FROM golang:1.22-bookworm AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=1 go build -o /capture-service ./cmd/capture-service

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates libwayland-dev libx11-dev && rm -rf /var/lib/apt/lists/*
COPY --from=builder /capture-service /usr/local/bin/capture-service
ENTRYPOINT ["/usr/local/bin/capture-service"]
