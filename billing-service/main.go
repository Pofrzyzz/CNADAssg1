package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	dbDSN := os.Getenv("DB_DSN")

	// Connect to the database
	db, err := sql.Open("mysql", dbDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize router
	r := mux.NewRouter()

	// Billing endpoints
	r.HandleFunc("/calculate-cost", func(w http.ResponseWriter, r *http.Request) { CalculateCostHandler(w, r, db) }).Methods("POST")
	r.HandleFunc("/generate-invoice", func(w http.ResponseWriter, r *http.Request) { GenerateInvoiceHandler(w, r, db) }).Methods("POST")
	r.HandleFunc("/invoices/{reservation_id}", func(w http.ResponseWriter, r *http.Request) { FetchInvoiceHandler(w, r, db) }).Methods("GET")
	r.HandleFunc("/invoices/user/{user_id}", func(w http.ResponseWriter, r *http.Request) { FetchInvoicesByUserHandler(w, r, db) }).Methods("GET")
	r.HandleFunc("/update-payment-status", func(w http.ResponseWriter, r *http.Request) { UpdatePaymentStatusHandler(w, r, db) }).Methods("PATCH")
	r.HandleFunc("/make-payment", func(w http.ResponseWriter, r *http.Request) { MakePaymentHandler(w, r, db) }).Methods("POST")

	// Start server
	log.Println("Billing service running on port 8082")
	log.Fatal(http.ListenAndServe(":8082", r))
}

// CalculateCostHandler calculates the cost of a rental
func CalculateCostHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var data struct {
		MembershipTier string  `json:"membership_tier"`
		Hours          float64 `json:"hours"`
		Discount       float64 `json:"discount"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Define base rates based on membership tier
	baseRate := map[string]float64{
		"Basic":   10.0,
		"Premium": 8.0,
		"VIP":     5.0,
	}

	// Calculate cost
	rate, exists := baseRate[data.MembershipTier]
	if !exists {
		http.Error(w, "Invalid membership tier", http.StatusBadRequest)
		return
	}
	cost := (rate * data.Hours) * ((100 - data.Discount) / 100)

	// Respond with the calculated cost
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]float64{"estimated_cost": cost})
}

// GenerateInvoiceHandler generates an invoice for a completed rental
func GenerateInvoiceHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var data struct {
		ReservationID int     `json:"reservation_id"`
		Amount        float64 `json:"amount"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	_, err = db.Exec(`INSERT INTO billing (reservation_id, amount, payment_status) VALUES (?, ?, ?)`,
		data.ReservationID, data.Amount, "Pending")
	if err != nil {
		http.Error(w, "Failed to generate invoice", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Invoice generated successfully"})
}

// FetchInvoiceHandler retrieves an invoice for a specific reservation
func FetchInvoiceHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	vars := mux.Vars(r)
	reservationID, err := strconv.Atoi(vars["reservation_id"])
	if err != nil {
		http.Error(w, "Invalid reservation ID", http.StatusBadRequest)
		return
	}

	var invoice struct {
		ReservationID int       `json:"reservation_id"`
		Amount        float64   `json:"amount"`
		PaymentStatus string    `json:"payment_status"`
		CreatedAt     time.Time `json:"created_at"`
	}

	err = db.QueryRow(`SELECT reservation_id, amount, payment_status, created_at FROM billing WHERE reservation_id = ?`, reservationID).
		Scan(&invoice.ReservationID, &invoice.Amount, &invoice.PaymentStatus, &invoice.CreatedAt)
	if err != nil {
		http.Error(w, "Invoice not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(invoice)
}

// UpdatePaymentStatusHandler updates the payment status of a reservation
func UpdatePaymentStatusHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var data struct {
		ReservationID int    `json:"reservation_id"`
		Status        string `json:"status"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	_, err = db.Exec(`UPDATE billing SET payment_status = ? WHERE reservation_id = ?`, data.Status, data.ReservationID)
	if err != nil {
		http.Error(w, "Failed to update payment status", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Payment status updated successfully"})
}

// FetchInvoicesByUserHandler retrieves all invoices for a specific user
func FetchInvoicesByUserHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["user_id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Query the database for invoices
	rows, err := db.Query(`
        SELECT 
            billing.reservation_id, 
            billing.amount, 
            billing.payment_status, 
            DATE_FORMAT(billing.created_at, '%Y-%m-%d') AS created_at 
        FROM billing
        JOIN reservations ON billing.reservation_id = reservations.reservation_id
        WHERE reservations.user_id = ?`, userID)
	if err != nil {
		http.Error(w, "Failed to fetch invoices", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Define a struct for individual invoice data
	var invoices []struct {
		ReservationID int     `json:"reservation_id"`
		Amount        float64 `json:"amount"`
		PaymentStatus string  `json:"payment_status"`
		CreatedAt     string  `json:"created_at"` // Use string for date
	}

	// Loop through the rows and scan data into the struct
	for rows.Next() {
		var invoice struct {
			ReservationID int     `json:"reservation_id"`
			Amount        float64 `json:"amount"`
			PaymentStatus string  `json:"payment_status"`
			CreatedAt     string  `json:"created_at"` // Use string for date
		}
		if err := rows.Scan(&invoice.ReservationID, &invoice.Amount, &invoice.PaymentStatus, &invoice.CreatedAt); err != nil {
			http.Error(w, "Error scanning invoice data", http.StatusInternalServerError)
			return
		}
		invoices = append(invoices, invoice)
	}

	// Check for errors during iteration
	if err := rows.Err(); err != nil {
		http.Error(w, "Error iterating over rows", http.StatusInternalServerError)
		return
	}

	// Encode the data to JSON and send it in the response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(invoices)
}

// MakePaymentHandler creates a new billing record for a payment
func MakePaymentHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var data struct {
		ReservationID int     `json:"reservation_id"`
		Amount        float64 `json:"amount"`
		PaymentStatus string  `json:"payment_status"`
		InvoiceID     string  `json:"invoice_id"`
	}

	// Decode the incoming request body
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if data.ReservationID == 0 || data.Amount <= 0 || data.PaymentStatus == "" || data.InvoiceID == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	// Insert a new row into the billing table
	_, err = db.Exec(`
		INSERT INTO billing (reservation_id, amount, payment_status, invoice_id, created_at)
		VALUES (?, ?, ?, ?, NOW())`,
		data.ReservationID, data.Amount, data.PaymentStatus, data.InvoiceID,
	)
	if err != nil {
		http.Error(w, "Failed to create billing record", http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Billing record created successfully"})
}
