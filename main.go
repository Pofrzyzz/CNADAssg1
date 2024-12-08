package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
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

	// Proxy routes to microservices
	r.HandleFunc("/api/user/{endpoint:.*}", proxyHandler("http://localhost:8081")).Methods("GET", "POST", "PATCH", "DELETE")
	r.HandleFunc("/api/billing/{endpoint:.*}", proxyHandler("http://localhost:8082")).Methods("GET", "POST", "PATCH", "DELETE")
	r.HandleFunc("/api/vehicle/{endpoint:.*}", proxyHandler("http://localhost:8083")).Methods("GET", "POST", "PATCH", "DELETE")

	// Debugging proxy
	r.HandleFunc("/api/debug", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Proxy is working")
	})

	// Serve static files
	staticDir := "./static"
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir(staticDir))))

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
	}{
		{"user-service", "user-service"},
		{"billing-service", "billing-service"},
		{"vehicle-service", "vehicle-service"},
	}

	for _, service := range services {
		cmd := exec.Command("go", "run", filepath.Join(service.Path, "main.go"))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Start()
		if err != nil {
			log.Fatalf("Failed to start %s: %v", service.Name, err)
		}
		log.Printf("%s started", service.Name)
	}
}

// proxyHandler correctly proxies requests to the target microservice
func proxyHandler(target string) func(w http.ResponseWriter, r *http.Request) {
	// Parse the target URL once
	fmt.Println("hello")
	targetURL, err := url.Parse(target)
	if err != nil {
		log.Fatalf("Invalid target URL %s: %v", target, err)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	return func(w http.ResponseWriter, r *http.Request) {
		// Update the request URL to match the target
		r.URL.Host = targetURL.Host
		r.URL.Scheme = targetURL.Scheme
		r.URL.Path = "/" + mux.Vars(r)["endpoint"]
		r.Host = targetURL.Host
		// fmt.Println(r.Host)
		// fmt.Println(r.URL.Path)
		// fmt.Println(r.URL.Host)

		// Debugging
		log.Printf("Proxying request: %s %s -> %s", r.Method, r.URL.Path, targetURL.String())

		// Use the reverse proxy to forward the request
		proxy.ServeHTTP(w, r)
	}
}
