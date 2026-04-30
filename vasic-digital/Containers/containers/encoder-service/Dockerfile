# vasic-digital/Containers/containers/encoder-service/Dockerfile
FROM golang:1.22-bookworm AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=1 go build -o /encoder-service ./cmd/encoder-service

FROM nvidia/cuda:12.4.0-base-ubuntu22.04
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /encoder-service /usr/local/bin/encoder-service
ENTRYPOINT ["/usr/local/bin/encoder-service"]
