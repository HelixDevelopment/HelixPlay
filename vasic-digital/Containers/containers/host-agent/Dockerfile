# vasic-digital/Containers/containers/host-agent/Dockerfile
FROM golang:1.22-bookworm AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o /host-agent ./cmd/host-agent

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /host-agent /usr/local/bin/host-agent
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/host-agent"]
