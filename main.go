package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"me.absen/app/db"
	"me.absen/app/handlers"
	"me.absen/app/static"
)

func main() {
	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		supabaseURL = "https://sqlxflkwmyogxmhdtlgu.supabase.co"
	}

	supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	if supabaseKey == "" {
		supabaseKey = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InNxbHhmbGt3bXlvZ3htaGR0bGd1Iiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImlhdCI6MTc4NDkxODE4NSwiZXhwIjoyMTAwNDk0MTg1fQ.7YkTWWijC1VjnH9pt4hlyCVNvKAmk7nBKy1x98tacQ0"
	}

	store := db.NewSupabaseStore(supabaseURL, supabaseKey)
	handler := handlers.NewServerHandler(store)

	mux := http.NewServeMux()

	// Page Routes (Separated Login & Dashboard Pages)
	mux.HandleFunc("/", handler.IndexPageHandler)
	mux.HandleFunc("/login", handler.LoginPageHandler)
	mux.HandleFunc("/dashboard", handler.DashboardPageHandler)

	// API Routes
	mux.HandleFunc("/api/login", handler.LoginHandler)
	mux.HandleFunc("/api/me", handler.MeHandler)
	mux.HandleFunc("/api/register-ip", handler.RegisterIPHandler)
	mux.HandleFunc("/api/absen", handler.AbsenHandler)
	mux.HandleFunc("/api/logs", handler.LogsHandler)
	mux.HandleFunc("/api/export-pdf", handler.ExportPDFHandler)
	mux.HandleFunc("/api/logout", handler.LogoutHandler)

	// Serve Static Embedded CSS/JS Assets
	fs := http.FileServer(http.FS(static.Files))
	mux.Handle("/style.css", fs)
	mux.Handle("/login.js", fs)
	mux.Handle("/dashboard.js", fs)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("====================================================\n")
	fmt.Printf("🚀 ABSENSI MAHASISWA MAGANG PT. ADYA ARTHA ABADI BALI\n")
	fmt.Printf("📍 TARGET LOKASI : Jalan Gatot Subroto I, Tonja, Denpasar\n")
	fmt.Printf("⏰ BATAS WAKTU  : 08:30:00 WITA\n")
	fmt.Printf("🔒 ANTI-TIPSEN  : Strict 1 Registered IP Per User\n")
	fmt.Printf("☁️ DATABASE     : Supabase Connected (%s)\n", supabaseURL)
	fmt.Printf("🔐 LOGIN PAGE   : http://localhost:%s/login\n", port)
	fmt.Printf("📊 DASHBOARD    : http://localhost:%s/dashboard\n", port)
	fmt.Printf("====================================================\n")

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server Error: %v", err)
	}
}
