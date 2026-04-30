# vasic-digital/Containers/containers/discovery-beacon/Dockerfile
FROM golang:1.22-bookworm AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o /discovery-beacon ./cmd/discovery-beacon

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /discovery-beacon /usr/local/bin/discovery-beacon
EXPOSE 5353/udp
ENTRYPOINT ["/usr/local/bin/discovery-beacon"]
