package handler

import (
	"net/http"
	"os"

	"me.absen/app/db"
	"me.absen/app/handlers"
	"me.absen/app/static"
)

var mux *http.ServeMux

func init() {
	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		supabaseURL = "https://sqlxflkwmyogxmhdtlgu.supabase.co"
	}

	supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	if supabaseKey == "" {
		supabaseKey = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InNxbHhmbGt3bXlvZ3htaGR0bGd1Iiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImlhdCI6MTc4NDkxODE4NSwiZXhwIjoyMTAwNDk0MTg1fQ.7YkTWWijC1VjnH9pt4hlyCVNvKAmk7nBKy1x98tacQ0"
	}

	store := db.NewSupabaseStore(supabaseURL, supabaseKey)
	serverHandler := handlers.NewServerHandler(store)

	mux = http.NewServeMux()

	// Page Routes
	mux.HandleFunc("/", serverHandler.IndexPageHandler)
	mux.HandleFunc("/login", serverHandler.LoginPageHandler)
	mux.HandleFunc("/dashboard", serverHandler.DashboardPageHandler)

	// API Routes
	mux.HandleFunc("/api/login", serverHandler.LoginHandler)
	mux.HandleFunc("/api/me", serverHandler.MeHandler)
	mux.HandleFunc("/api/register-ip", serverHandler.RegisterIPHandler)
	mux.HandleFunc("/api/absen", serverHandler.AbsenHandler)
	mux.HandleFunc("/api/logs", serverHandler.LogsHandler)
	mux.HandleFunc("/api/logout", serverHandler.LogoutHandler)

	// Serve Static Embedded Assets
	fs := http.FileServer(http.FS(static.Files))
	mux.Handle("/style.css", fs)
	mux.Handle("/login.js", fs)
	mux.Handle("/dashboard.js", fs)
}

// Handler is the Vercel serverless function entrypoint
func Handler(w http.ResponseWriter, r *http.Request) {
	mux.ServeHTTP(w, r)
}
