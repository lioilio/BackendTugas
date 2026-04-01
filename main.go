package main

import (
    "log"
    "net/http"
    "os"
)

const maxUploadSize = 10 << 20 // 10 MiB
const uploadDir = "uploads"

func main() {
    // Create upload directory if it doesn't exist.
    if err := os.MkdirAll(uploadDir, 0755); err != nil {
        log.Fatalf("Could not create upload directory %q: %v", uploadDir, err)
    }

    // Register upload handler path
    http.HandleFunc("/upload", uploadHandler)

    addr := ":8080"
    log.Printf("Starting server on %s...", addr)
    if err := http.ListenAndServe(addr, nil); err != nil {
        log.Fatalf("Server failed: %v", err)
    }
}
