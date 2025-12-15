package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/rs/cors"

	"go-template/renderer"
)

func handleRender(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req renderer.RenderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, renderer.RenderResponse{
			Error: "Invalid request body: " + err.Error(),
		})
		return
	}

	resp := renderer.Render(req)
	if resp.Error != "" {
		respondJSON(w, http.StatusBadRequest, resp)
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func respondJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

func main() {
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		ExposedHeaders:   []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           300,
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/api/render", handleRender)
	mux.HandleFunc("/health", handleHealth)

	handler := c.Handler(mux)

	addr := ":8080"
	log.Printf("Starting server on %s", addr)
	log.Printf("POST http://localhost:8080/api/render")
	log.Printf("GET  http://localhost:8080/health")

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
