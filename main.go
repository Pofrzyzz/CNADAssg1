package main

import (
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Start microservices
	startMicroservices()

	// Initialize router
	r := mux.NewRouter()

	// Serve static files
	staticDir := "./static"
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir(staticDir))))

	// Proxy routes to microservices
	r.HandleFunc("/api/user/{endpoint:.*}", proxyHandler("http://localhost:8081")).Methods("GET", "POST", "PATCH", "DELETE")
	r.HandleFunc("/api/billing/{endpoint:.*}", proxyHandler("http://localhost:8082")).Methods("GET", "POST", "PATCH", "DELETE")
	r.HandleFunc("/api/vehicle/{endpoint:.*}", proxyHandler("http://localhost:8083")).Methods("GET", "POST", "PATCH", "DELETE")

	// Start the server
	port := os.Getenv("MAIN_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Main service running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

// startMicroservices starts all the microservices as subprocesses
func startMicroservices() {
	services := []struct {
		Name string
		Path string
		Port string
	}{
		{"user-service", "user-service", "8081"},
		{"billing-service", "billing-service", "8082"},
		{"vehicle-service", "vehicle-service", "8083"},
	}

	for _, service := range services {
		cmd := exec.Command("go", "run", filepath.Join(service.Path, "main.go"))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Start()
		if err != nil {
			log.Fatalf("Failed to start %s: %v", service.Name, err)
		}
		log.Printf("%s started on port %s", service.Name, service.Port)
	}
}

// proxyHandler redirects requests to the corresponding microservice
func proxyHandler(target string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		proxy := http.NewServeMux()
		proxy.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			url := target + "/" + mux.Vars(req)["endpoint"]
			http.Redirect(w, req, url, http.StatusTemporaryRedirect)
		}))
		proxy.ServeHTTP(w, r)
	}
}
