package main

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
)

type jsonResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
    File    string `json:"file,omitempty"`
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
    // Ensure only POST is accepted.
    if r.Method != http.MethodPost {
        respondJSON(w, http.StatusMethodNotAllowed, jsonResponse{
            Success: false,
            Message: "Invalid method: only POST is allowed",
        })
        return
    }

    // Constrain request body size.
    r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize+1024)
    if err := r.ParseMultipartForm(maxUploadSize); err != nil {
        if err.Error() == "http: request body too large" {
            respondJSON(w, http.StatusRequestEntityTooLarge, jsonResponse{
                Success: false,
                Message: "File too large. Maximum allowed is 10 MB",
            })
            return
        }
        respondJSON(w, http.StatusBadRequest, jsonResponse{
            Success: false,
            Message: "Invalid multipart/form-data: " + err.Error(),
        })
        return
    }

    // Get file from form field named "file".
    file, fileHeader, err := r.FormFile("file")
    if err != nil {
        respondJSON(w, http.StatusBadRequest, jsonResponse{
            Success: false,
            Message: "Missing 'file' in form data: " + err.Error(),
        })
        return
    }
    defer file.Close()

    filename := filepath.Base(fileHeader.Filename)
    if filename == "" || filename == "." || filename == "/" {
        respondJSON(w, http.StatusBadRequest, jsonResponse{
            Success: false,
            Message: "Invalid filename",
        })
        return
    }

    // Ensure upload directory exists.
    if err := os.MkdirAll(uploadDir, 0755); err != nil {
        respondJSON(w, http.StatusInternalServerError, jsonResponse{
            Success: false,
            Message: "Could not create upload directory: " + err.Error(),
        })
        return
    }

    dstPath := filepath.Join(uploadDir, filename)

    // Try to avoid overwrite by appending process id if already exists.
    if _, err := os.Stat(dstPath); err == nil {
        dstPath = filepath.Join(uploadDir, fmt.Sprintf("%s_%d", filename, os.Getpid()))
    }

    outFile, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
    if err != nil {
        if os.IsExist(err) {
            respondJSON(w, http.StatusConflict, jsonResponse{
                Success: false,
                Message: "File already exists",
            })
            return
        }
        respondJSON(w, http.StatusInternalServerError, jsonResponse{
            Success: false,
            Message: "Could not create destination file: " + err.Error(),
        })
        return
    }
    defer outFile.Close()

    if _, err := io.Copy(outFile, file); err != nil {
        respondJSON(w, http.StatusInternalServerError, jsonResponse{
            Success: false,
            Message: "Error writing file: " + err.Error(),
        })
        return
    }

    respondJSON(w, http.StatusOK, jsonResponse{
        Success: true,
        Message: "File uploaded successfully",
        File:    dstPath,
    })
}

func respondJSON(w http.ResponseWriter, status int, data jsonResponse) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(data)
}
