package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

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

	// Vehicle endpoints
	r.HandleFunc("/vehicles", func(w http.ResponseWriter, r *http.Request) { GetVehiclesHandler(w, r, db) }).Methods("GET")
	r.HandleFunc("/book-vehicle", func(w http.ResponseWriter, r *http.Request) { BookVehicleHandler(w, r, db) }).Methods("POST")
	r.HandleFunc("/modify-booking", func(w http.ResponseWriter, r *http.Request) { ModifyBookingHandler(w, r, db) }).Methods("PATCH")
	r.HandleFunc("/cancel-booking", func(w http.ResponseWriter, r *http.Request) { CancelBookingHandler(w, r, db) }).Methods("DELETE")
	r.HandleFunc("/vehicle-status/{vehicle_id}", func(w http.ResponseWriter, r *http.Request) { GetVehicleStatusHandler(w, r, db) }).Methods("GET")
	r.HandleFunc("/find-reservationid", func(w http.ResponseWriter, r *http.Request) { FindReservationIDHandler(w, r, db) }).Methods("POST")
	r.HandleFunc("/update-history", func(w http.ResponseWriter, r *http.Request) { UpdateHistoryHandler(w, r, db) }).Methods("POST")
	r.HandleFunc("/retrieve-model", func(w http.ResponseWriter, r *http.Request) { RetrieveModelHandler(w, r, db) }).Methods("POST")
	r.HandleFunc("/retrieve-vehid", func(w http.ResponseWriter, r *http.Request) { RetrieveVehicleIDHandler(w, r, db) }).Methods("POST")

	// Start server
	log.Println("Vehicle service running on port 8083")
	log.Fatal(http.ListenAndServe(":8083", r))
}

// GetVehiclesHandler retrieves all vehicles with optional filters
func GetVehiclesHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	query := `SELECT vehicle_id, model, license_plate, charge_level, availability, location FROM vehicles WHERE 1=1`

	// Filter by availability if provided
	availability := r.URL.Query().Get("availability")
	if availability != "" {
		query += " AND availability = " + availability
	}

	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, "Failed to fetch vehicles", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var vehicles []struct {
		VehicleID    int    `json:"vehicle_id"`
		Model        string `json:"model"`
		LicensePlate string `json:"license_plate"`
		ChargeLevel  int    `json:"charge_level"`
		Availability bool   `json:"availability"`
		Location     string `json:"location"`
	}

	for rows.Next() {
		var vehicle struct {
			VehicleID    int    `json:"vehicle_id"`
			Model        string `json:"model"`
			LicensePlate string `json:"license_plate"`
			ChargeLevel  int    `json:"charge_level"`
			Availability bool   `json:"availability"`
			Location     string `json:"location"`
		}
		if err := rows.Scan(&vehicle.VehicleID, &vehicle.Model, &vehicle.LicensePlate, &vehicle.ChargeLevel, &vehicle.Availability, &vehicle.Location); err != nil {
			http.Error(w, "Error scanning vehicle data", http.StatusInternalServerError)
			return
		}
		vehicles = append(vehicles, vehicle)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(vehicles)
}

// BookVehicleHandler books a vehicle for a specified time range
func BookVehicleHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var data struct {
		UserID    int    `json:"user_id"`
		VehicleID int    `json:"vehicle_id"`
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Check if the vehicle is available
	var availability bool
	err = db.QueryRow(`SELECT availability FROM vehicles WHERE vehicle_id = ?`, data.VehicleID).Scan(&availability)
	if err != nil || !availability {
		http.Error(w, "Vehicle not available", http.StatusConflict)
		return
	}

	// Create a reservation
	_, err = db.Exec(`INSERT INTO reservations (user_id, vehicle_id, start_time, end_time, status, created_at) VALUES (?, ?, ?, ?, 'Booked', NOW())`,
		data.UserID, data.VehicleID, data.StartTime, data.EndTime)
	if err != nil {
		http.Error(w, "Failed to book vehicle", http.StatusInternalServerError)
		return
	}

	// Mark the vehicle as unavailable
	_, err = db.Exec(`UPDATE vehicles SET availability = FALSE WHERE vehicle_id = ?`, data.VehicleID)
	if err != nil {
		http.Error(w, "Failed to update vehicle availability", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Vehicle booked successfully"})
}

// ModifyBookingHandler modifies an existing reservation
func ModifyBookingHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var data struct {
		ReservationID int    `json:"reservation_id"`
		StartTime     string `json:"start_time"`
		EndTime       string `json:"end_time"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Update the reservation
	_, err = db.Exec(`UPDATE reservations SET start_time = ?, end_time = ? WHERE reservation_id = ? AND status = 'Booked'`,
		data.StartTime, data.EndTime, data.ReservationID)
	if err != nil {
		http.Error(w, "Failed to modify booking", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Booking modified successfully"})
}

// CancelBookingHandler cancels a reservation by removing entries from billing and reservations and updates vehicle availability
func CancelBookingHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var data struct {
		ReservationID int `json:"reservation_id"`
	}

	// Decode the incoming request body
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Get the vehicle ID associated with the reservation
	var vehicleID int
	err = db.QueryRow(`SELECT vehicle_id FROM reservations WHERE reservation_id = ?`, data.ReservationID).Scan(&vehicleID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Reservation not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to fetch reservation details", http.StatusInternalServerError)
		}
		return
	}

	// Remove the billing entry for the reservation
	_, err = db.Exec(`DELETE FROM billing WHERE reservation_id = ?`, data.ReservationID)
	if err != nil {
		http.Error(w, "Failed to delete billing record", http.StatusInternalServerError)
		return
	}

	// Remove the reservation entry
	_, err = db.Exec(`DELETE FROM reservations WHERE reservation_id = ?`, data.ReservationID)
	if err != nil {
		http.Error(w, "Failed to delete reservation record", http.StatusInternalServerError)
		return
	}

	// Mark the vehicle as available again
	_, err = db.Exec(`UPDATE vehicles SET availability = TRUE WHERE vehicle_id = ?`, vehicleID)
	if err != nil {
		http.Error(w, "Failed to update vehicle availability", http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Reservation and billing records cancelled successfully"})
}

// GetVehicleStatusHandler retrieves the status of a specific vehicle
func GetVehicleStatusHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	vars := mux.Vars(r)
	vehicleID, err := strconv.Atoi(vars["vehicle_id"])
	if err != nil {
		http.Error(w, "Invalid vehicle ID", http.StatusBadRequest)
		return
	}

	var vehicle struct {
		VehicleID   int    `json:"vehicle_id"`
		Model       string `json:"model"`
		ChargeLevel int    `json:"charge_level"`
		Location    string `json:"location"`
	}

	err = db.QueryRow(`SELECT vehicle_id, model, charge_level, location FROM vehicles WHERE vehicle_id = ?`, vehicleID).
		Scan(&vehicle.VehicleID, &vehicle.Model, &vehicle.ChargeLevel, &vehicle.Location)
	if err != nil {
		http.Error(w, "Vehicle not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(vehicle)
}

// FindReservationIDHandler retrieves the reservation_id based on user_id, vehicle_id, start_time, and end_time from JSON input
func FindReservationIDHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Define the expected input structure
	var input struct {
		UserID    int    `json:"user_id"`
		VehicleID int    `json:"vehicle_id"`
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
	}

	// Decode the JSON body
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid JSON input", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if input.UserID == 0 || input.VehicleID == 0 || input.StartTime == "" || input.EndTime == "" {
		http.Error(w, "All fields (user_id, vehicle_id, start_time, end_time) are required", http.StatusBadRequest)
		return
	}

	// Query to fetch reservation_id
	var reservationID int
	err = db.QueryRow(`
        SELECT reservation_id 
        FROM reservations 
        WHERE user_id = ? AND vehicle_id = ? AND start_time = ? AND end_time = ?`,
		input.UserID, input.VehicleID, input.StartTime, input.EndTime).Scan(&reservationID)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Reservation not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to fetch reservation ID", http.StatusInternalServerError)
		}
		return
	}

	// Respond with the reservation ID
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int{"reservation_id": reservationID})
}

// UpdateHistoryHandler updates the user's rental history with a new booking
func UpdateHistoryHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var data struct {
		UserID    int     `json:"user_id"`
		VehicleID int     `json:"vehicle_id"`
		StartTime string  `json:"start_time"`
		EndTime   string  `json:"end_time"`
		TotalCost float64 `json:"total_cost"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	fmt.Println(err)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Insert rental history record into the database
	_, err = db.Exec(`
        INSERT INTO rental_history (user_id, vehicle_id, start_time, end_time, total_cost)
        VALUES (?, ?, ?, ?, ?)`,
		data.UserID, data.VehicleID, data.StartTime, data.EndTime, data.TotalCost)
	if err != nil {
		http.Error(w, "Failed to update rental history", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Rental history updated successfully"})
}

// RetrieveModelHandler retrieves the vehicle model based on the vehicle ID
func RetrieveModelHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Define the input structure
	var input struct {
		VehicleID int `json:"vehicle_id"`
	}

	// Decode the JSON body
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid JSON input", http.StatusBadRequest)
		return
	}

	// Validate the input
	if input.VehicleID == 0 {
		http.Error(w, "Vehicle ID is required", http.StatusBadRequest)
		return
	}

	// Query the database for the vehicle model
	var vehicleModel string
	err = db.QueryRow(`SELECT model FROM vehicles WHERE vehicle_id = ?`, input.VehicleID).Scan(&vehicleModel)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Vehicle not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve vehicle model", http.StatusInternalServerError)
		}
		return
	}

	// Return the vehicle model in the response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"vehicle_model": vehicleModel})
}

// RetrieveVehicleIDHandler retrieves the vehicle_id based on the vehicle model
func RetrieveVehicleIDHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Define the input structure
	var input struct {
		VehicleModel string `json:"vehicle_model"`
	}

	// Decode the JSON body
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid JSON input", http.StatusBadRequest)
		return
	}

	// Validate the input
	if input.VehicleModel == "" {
		http.Error(w, "Vehicle model is required", http.StatusBadRequest)
		return
	}

	// Query the database for the vehicle ID
	var vehicleID int
	err = db.QueryRow(`SELECT vehicle_id FROM vehicles WHERE model = ?`, input.VehicleModel).Scan(&vehicleID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Vehicle not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve vehicle ID", http.StatusInternalServerError)
		}
		return
	}

	// Return the vehicle ID in the response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int{"vehicle_id": vehicleID})
}
