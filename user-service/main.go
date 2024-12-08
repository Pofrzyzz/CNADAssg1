package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/go-sql-driver/mysql"
)

// JWT secret key
var jwtKey []byte

// Temporary OTP storage
var otpStore = make(map[string]string)

// Claims structure for JWT
type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// User struct for profile data
type User struct {
	Email          string `json:"email"`
	PhoneNumber    string `json:"phone_number"`
	CountryCode    string `json:"country_code"`
	MembershipTier string `json:"membership_tier"`
	PhoneVerified  bool   `json:"phone_verified"`
}

// RentalHistory struct
type RentalHistory struct {
	VehicleModel string  `json:"vehicle_model"`
	StartTime    string  `json:"start_time"`
	EndTime      string  `json:"end_time"`
	TotalCost    float64 `json:"total_cost"`
}

func main() {
	// Load environment variables
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	dbDSN := os.Getenv("DB_DSN")
	jwtKey = []byte(os.Getenv("JWT_SECRET_KEY"))

	// Connect to the database
	db, err := sql.Open("mysql", dbDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize router
	r := mux.NewRouter()

	// Phone verification endpoints
	r.HandleFunc("/generate-otp", GenerateOTPHandler).Methods("POST")
	r.HandleFunc("/verify-otp", func(w http.ResponseWriter, r *http.Request) { VerifyOTPHandler(w, r, db) }).Methods("POST")

	// User management endpoints
	r.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) { RegisterHandler(w, r, db) }).Methods("POST")
	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) { LoginHandler(w, r, db) }).Methods("POST")
	r.HandleFunc("/profile", func(w http.ResponseWriter, r *http.Request) { ProfileHandler(w, r, db) }).Methods("GET")
	r.HandleFunc("/user-id", func(w http.ResponseWriter, r *http.Request) { GetUserIDHandler(w, r, db) }).Methods("GET")

	// Rental history endpoint
	r.HandleFunc("/rental-history", func(w http.ResponseWriter, r *http.Request) { RentalHistoryHandler(w, r, db) }).Methods("GET")

	// Start server
	log.Println("User service running on port 8081")
	log.Fatal(http.ListenAndServe(":8081", r))
}

// GenerateOTPHandler generates a random OTP
func GenerateOTPHandler(w http.ResponseWriter, r *http.Request) {
	phone := r.FormValue("phone_number")
	if phone == "" {
		http.Error(w, "Phone number is required", http.StatusBadRequest)
		return
	}

	otp := fmt.Sprintf("%06d", rand.Intn(1000000))
	otpStore[phone] = otp

	fmt.Printf("OTP for %s: %s\n", phone, otp)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "OTP sent successfully"})
}

// VerifyOTPHandler verifies the OTP
func VerifyOTPHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	phone := r.FormValue("phone_number")
	otp := r.FormValue("otp")

	if otpStore[phone] != otp {
		http.Error(w, "Invalid OTP", http.StatusUnauthorized)
		return
	}

	delete(otpStore, phone)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Phone verified successfully"})
}

// RegisterHandler registers a new user
func RegisterHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	phone := r.FormValue("phone_number")
	countryCode := r.FormValue("country_code")

	if _, exists := otpStore[phone]; exists {
		http.Error(w, "Phone number not verified", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec(`INSERT INTO users (email, password_hash, phone_number, country_code, phone_verified) VALUES (?, ?, ?, ?, TRUE)`,
		email, hashedPassword, phone, countryCode)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}

// LoginHandler handles user login
func LoginHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	var hashedPassword string
	err := db.QueryRow(`SELECT password_hash FROM users WHERE email = ?`, email).Scan(&hashedPassword)
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	token, err := generateJWT(email)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

// ProfileHandler fetches the user profile
func ProfileHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	authHeader := r.Header.Get("Authorization")
	tokenString := authHeader[len("Bearer "):]

	claims, err := validateJWT(tokenString)
	if err != nil {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	var user User
	err = db.QueryRow(`SELECT email, phone_number, country_code, membership_tier, phone_verified FROM users WHERE email = ?`, claims.Email).
		Scan(&user.Email, &user.PhoneNumber, &user.CountryCode, &user.MembershipTier, &user.PhoneVerified)
	fmt.Println(err)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

// GetUserIDHandler retrieves the user_id of the logged-in user
func GetUserIDHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	authHeader := r.Header.Get("Authorization")
	tokenString := authHeader[len("Bearer "):]

	claims, err := validateJWT(tokenString)
	if err != nil {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	var userID int
	err = db.QueryRow(`SELECT user_id FROM users WHERE email = ?`, claims.Email).Scan(&userID)
	log.Println(claims)
	log.Println(claims.Email)
	log.Println(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int{"user_id": userID})
}

// RentalHistoryHandler retrieves rental history
func RentalHistoryHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	authHeader := r.Header.Get("Authorization")
	tokenString := authHeader[len("Bearer "):]
	fmt.Println(tokenString)
	claims, err := validateJWT(tokenString)
	if err != nil {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	rows, err := db.Query(`SELECT vehicles.model, rental_history.start_time, rental_history.end_time, rental_history.total_cost
		FROM rental_history
		JOIN users ON users.user_id = rental_history.user_id
		JOIN vehicles ON vehicles.vehicle_id = rental_history.vehicle_id
		WHERE users.email = ?`, claims.Email)
	if err != nil {
		http.Error(w, "Failed to fetch rental history", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var rentals []RentalHistory
	for rows.Next() {
		var rental RentalHistory
		if err := rows.Scan(&rental.VehicleModel, &rental.StartTime, &rental.EndTime, &rental.TotalCost); err != nil {
			http.Error(w, "Error scanning rental history", http.StatusInternalServerError)
			return
		}
		rentals = append(rentals, rental)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(rentals)
}

// JWT utilities
func generateJWT(email string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func validateJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	return claims, nil
}
