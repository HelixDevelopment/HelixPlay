module github.com/HelixDevelopment/HelixPlay

go 1.26.2

require (
	digital.vasic.auth v0.0.0-00010101000000-000000000000
	github.com/0xcafed00d/joystick v1.0.1
	github.com/HelixDevelopment/HelixPlay/vasic-digital/Memory v0.0.0-00010101000000-000000000000
	github.com/andybalholm/brotli v1.2.1
	github.com/hashicorp/mdns v1.0.6
	google.golang.org/grpc v1.80.0
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/golang-jwt/jwt/v5 v5.2.2 // indirect
	github.com/miekg/dns v1.1.55 // indirect
	go.opentelemetry.io/otel v1.43.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.43.0 // indirect
	golang.org/x/mod v0.34.0 // indirect
	golang.org/x/net v0.52.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	golang.org/x/tools v0.43.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260401024825-9d38bb4040a9 // indirect
)

replace (
	digital.vasic.auth => ./Auth
	github.com/HelixDevelopment/HelixPlay/vasic-digital/Memory => ./Memory
)
