module github.com/HelixDevelopment/HelixPlay

go 1.26.2

require (
	github.com/HelixDevelopment/HelixPlay/vasic-digital/Memory v0.0.0-00010101000000-000000000000
	github.com/andybalholm/brotli v1.2.1
	github.com/google/uuid v1.6.0
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/HelixDevelopment/HelixPlay/vasic-digital/Memory => ./vasic-digital/Memory
