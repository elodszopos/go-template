package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"regexp"
    "strconv"

	"github.com/rs/cors"
)


// RenderRequest represents the incoming request payload
type RenderRequest struct {
	Template string          `json:"template"`
	Data     json.RawMessage `json:"data"`
}

type RenderResponse struct {
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
	Line   int    `json:"line,omitempty"`
	Column int    `json:"column,omitempty"`
}

// handleRender processes template rendering requests
func handleRender(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse incoming request
	var req RenderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, RenderResponse{
			Error: "Invalid request body: " + err.Error(),
		})
		return
	}

	// Parse the template string
	tmpl, err := template.New("template").Parse(req.Template)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, RenderResponse{
			Error: "Template parse error: " + err.Error(),
		})
		return
	}

	// Unmarshal data into a generic interface
	var data interface{}
	if err := json.Unmarshal(req.Data, &data); err != nil {
		respondJSON(w, http.StatusBadRequest, RenderResponse{
			Error: "Data parse error: " + err.Error(),
		})
		return
	}

	// Execute the template
    	var buf bytes.Buffer
    	if err := tmpl.Execute(&buf, data); err != nil {
    		// Try to extract line number from error
    		errMsg := err.Error()
    		line, column := extractLineColumn(errMsg)

    		respondJSON(w, http.StatusBadRequest, RenderResponse{
    			Error:  errMsg,
    			Line:   line,
    			Column: column,
    		})
    		return
    	}

	// Return the rendered output
	respondJSON(w, http.StatusOK, RenderResponse{
		Output: buf.String(),
	})
}

func extractLineColumn(errMsg string) (int, int) {
	// Matches patterns like "template:1:10:" for line 1, column 10
	re := regexp.MustCompile(`template:(\d+):(\d+):`)
	matches := re.FindStringSubmatch(errMsg)
	if len(matches) > 2 {
		line, _ := strconv.Atoi(matches[1])
		col, _ := strconv.Atoi(matches[2])
		return line, col
	}
	return 0, 0
}

// respondJSON writes a JSON response
func respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// handleHealth is a simple health check endpoint
func handleHealth(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func main() {
	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		ExposedHeaders:   []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           300,
	})

	// Setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/api/render", handleRender)
	mux.HandleFunc("/health", handleHealth)

	// Wrap with CORS middleware
	handler := c.Handler(mux)

	// Start server
	addr := ":8080"
	log.Printf("Starting server on %s", addr)
	log.Printf("Template rendering endpoint: POST http://localhost:8080/api/render")
	log.Printf("Health check: GET http://localhost:8080/health")

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
