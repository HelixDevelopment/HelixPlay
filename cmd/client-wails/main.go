package main

import (
    "github.com/HelixDevelopment/HelixPlay/cmd/client-wails/backend"
)

func main() {
    b := backend.NewBackend()
    b.Start()
    // Wails.Run() would go here
    b.Stop()
}
