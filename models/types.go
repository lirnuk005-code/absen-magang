package models

import "time"

// User represents an account profile
type User struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	RegisteredIP string `json:"registered_ip"`
}

// AttendanceLog represents a record of check-in / check-out
type AttendanceLog struct {
	ID             string    `json:"id,omitempty"`
	Username       string    `json:"username"`
	CheckInTime    time.Time `json:"check_in_time"`
	Type           string    `json:"type"` // "DATANG" or "PULANG"
	Latitude       float64   `json:"latitude"`
	Longitude      float64   `json:"longitude"`
	DistanceMeters float64   `json:"distance_meters"`
	IPAddress      string    `json:"ip_address"`
	Status         string    `json:"status"` // SUCCESS, PULANG_NORMAL, PULANG_CEPAT, DITOLAK_WAKTU, DITOLAK_LOKASI, DITOLAK_IP_TIPSEN
	EarlyReason    string    `json:"early_reason,omitempty"`
	Notes          string    `json:"notes"`
}

// AbsenRequest represents payload sent from client browser
type AbsenRequest struct {
	Type        string  `json:"type"` // "DATANG" or "PULANG"
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	EarlyReason string  `json:"early_reason,omitempty"`
}

// LoginRequest payload
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Response generic wrapper
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
