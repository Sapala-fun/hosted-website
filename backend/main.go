package main

import (
    "encoding/json"
    "log"
    "net/http"
    "os"
    "path/filepath"

    "github.com/example/ownerrez-github-pages/internal/ownerrez"
)

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent)
            return
        }

        next(w, r)
    }
}

func corsHandler(client *ownerrez.Client, handler func(http.ResponseWriter, *http.Request, *ownerrez.Client)) http.HandlerFunc {
    return corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
        handler(w, r, client)
    })
}

func main() {
    client := ownerrez.NewClient()

    mux := http.NewServeMux()
    mux.HandleFunc("/api/health", corsMiddleware(healthHandler))
    mux.HandleFunc("/api/properties", corsHandler(client, propertiesHandler))
    mux.HandleFunc("/api/book", corsHandler(client, bookHandler))

    // Serve static files from the repository root (parent of backend/)
    cwd, _ := os.Getwd()
    repoRoot := filepath.Dir(cwd)
    fileServer := http.FileServer(http.Dir(repoRoot))
    apiMux := http.NewServeMux()
    apiMux.Handle("/api/", mux)
    apiMux.Handle("/", fileServer)

    port := os.Getenv("PORT")
    if port == "" {
        port = "3001"
    }

    log.Printf("server listening on :%s (frontend: http://localhost:%s)", port, port)
    if err := http.ListenAndServe(":"+port, apiMux); err != nil {
        log.Fatalf("server failed: %v", err)
    }
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(map[string]any{
        "status":    "ok",
        "service":   "ownerrez-proxy-go",
        "timestamp": "now",
    })
}

func propertiesHandler(w http.ResponseWriter, r *http.Request, client *ownerrez.Client) {
    w.Header().Set("Content-Type", "application/json")

    properties, err := client.GetProperties()
    if err != nil {
        _ = json.NewEncoder(w).Encode(map[string]any{
            "properties": []map[string]any{{
                "id":          "sample-property",
                "name":        "Ocean View Retreat",
                "slug":        "ocean-view-retreat",
                "nightlyRate": 210,
                "bedrooms":    2,
                "bathrooms":   2,
                "description": "Fallback property payload served by Go.",
            }},
            "warning": err.Error(),
        })
        return
    }

    _ = json.NewEncoder(w).Encode(map[string]any{"properties": properties})
}

func bookHandler(w http.ResponseWriter, r *http.Request, client *ownerrez.Client) {
    if r.Method != http.MethodPost {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var payload map[string]string
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        http.Error(w, "invalid json", http.StatusBadRequest)
        return
    }

    result, err := client.CreateBooking(payload)
    if err != nil {
        w.Header().Set("Content-Type", "application/json")
        _ = json.NewEncoder(w).Encode(map[string]any{
            "message": "Booking request received",
            "guestName": payload["guestName"],
            "checkIn": payload["checkIn"],
            "checkOut": payload["checkOut"],
            "note": err.Error(),
        })
        return
    }

    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(result)
}
