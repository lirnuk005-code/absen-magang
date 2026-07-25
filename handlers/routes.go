package handlers

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"me.absen/app/db"
	"me.absen/app/models"
	"me.absen/app/services"
	"me.absen/app/static"
)

type ServerHandler struct {
	Store *db.SupabaseStore
}

func NewServerHandler(store *db.SupabaseStore) *ServerHandler {
	return &ServerHandler{Store: store}
}

func ExtractClientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	xRealIP := r.Header.Get("X-Real-IP")
	if xRealIP != "" {
		return strings.TrimSpace(xRealIP)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// Page Routes
func (h *ServerHandler) IndexPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	cookie, err := r.Cookie("user_session")
	if err == nil && cookie.Value != "" {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (h *ServerHandler) LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("user_session")
	if err == nil && cookie.Value != "" {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}

	content, err := static.Files.ReadFile("login.html")
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(content)
}

func (h *ServerHandler) DashboardPageHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("user_session")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	content, err := static.Files.ReadFile("dashboard.html")
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(content)
}

// API Routes
func (h *ServerHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != "POST" {
		http.Error(w, `{"success":false,"message":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(models.Response{Success: false, Message: "Payload tidak valid"})
		return
	}

	user, err := h.Store.GetUser(req.Username)
	if err != nil || user.Password != req.Password {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(models.Response{Success: false, Message: "Username atau password salah!"})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "user_session",
		Value:    user.Username,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})

	clientIP := ExtractClientIP(r)
	json.NewEncoder(w).Encode(models.Response{
		Success: true,
		Message: "Login Berhasil",
		Data: map[string]interface{}{
			"username":      user.Username,
			"registered_ip": user.RegisteredIP,
			"current_ip":    clientIP,
		},
	})
}

type M map[string]interface{}

func (h *ServerHandler) MeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	cookie, err := r.Cookie("user_session")
	if err != nil || cookie.Value == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(models.Response{Success: false, Message: "Belum login"})
		return
	}

	user, err := h.Store.GetUser(cookie.Value)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(models.Response{Success: false, Message: "User tidak ditemukan"})
		return
	}

	clientIP := ExtractClientIP(r)
	now := time.Now().In(time.FixedZone("WITA", 8*3600))
	weekday := now.Weekday()
	isSunday := (weekday == time.Sunday)
	isPulangEarly := services.IsPulangEarly(weekday, now.Hour(), now.Minute())
	normalPulangHour := services.GetNormalPulangHour(weekday)
	dayName := services.GetDayNameIndonesian(weekday)

	json.NewEncoder(w).Encode(models.Response{
		Success: true,
		Data: M{
			"username":           user.Username,
			"registered_ip":      user.RegisteredIP,
			"current_ip":         clientIP,
			"target_loc":         "Jalan Gatot Subroto I, Tonja, Denpasar Utara, Denpasar, Bali",
			"target_lat":         services.TargetLat,
			"target_lng":         services.TargetLng,
			"day_name":           dayName,
			"is_sunday":          isSunday,
			"is_pulang_early":    isPulangEarly,
			"normal_pulang_hour": normalPulangHour,
		},
	})
}

func (h *ServerHandler) RegisterIPHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != "POST" {
		http.Error(w, `{"success":false,"message":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("user_session")
	if err != nil || cookie.Value == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(models.Response{Success: false, Message: "Silakan login terlebih dahulu"})
		return
	}

	clientIP := ExtractClientIP(r)
	err = h.Store.RegisterIP(cookie.Value, clientIP)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.Response{Success: false, Message: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(models.Response{
		Success: true,
		Message: fmt.Sprintf("IP %s Berhasil Didaftarkan Secara Permanen!", clientIP),
		Data: M{
			"registered_ip": clientIP,
		},
	})
}

func (h *ServerHandler) AbsenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != "POST" {
		http.Error(w, `{"success":false,"message":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("user_session")
	if err != nil || cookie.Value == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(models.Response{Success: false, Message: "Silakan login terlebih dahulu"})
		return
	}

	user, err := h.Store.GetUser(cookie.Value)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(models.Response{Success: false, Message: "User tidak ditemukan"})
		return
	}

	var req models.AbsenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(models.Response{Success: false, Message: "Payload presensi tidak valid"})
		return
	}

	if req.Type == "" {
		req.Type = "DATANG"
	}

	clientIP := ExtractClientIP(r)
	now := time.Now().In(time.FixedZone("WITA", 8*3600)) // Asia/Makassar WITA (UTC+8)
	weekday := now.Weekday()

	// 1. HARI MINGGU (AUTOMATED HOLIDAY RECORD)
	if weekday == time.Sunday {
		log := &models.AttendanceLog{
			Username:       user.Username,
			CheckInTime:    now,
			Type:           "LIBUR",
			Latitude:       req.Latitude,
			Longitude:      req.Longitude,
			DistanceMeters: 0,
			IPAddress:      clientIP,
			Status:         "LIBUR_HARI_MINGGU",
			Notes:          "Hari Minggu Libur (Terabsen Otomatis Libur oleh Sistem)",
		}
		h.Store.SaveAttendanceLog(log)

		json.NewEncoder(w).Encode(models.Response{
			Success: true,
			Message: "🌴 Hari ini Hari Minggu (Libur Kantor). Presensi Anda otomatis terverifikasi LIBUR!",
			Data:    log,
		})
		return
	}

	// 2. Check IP Registration (Anti-Tipsen)
	if user.RegisteredIP == "" {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(models.Response{
			Success: false,
			Message: "Anda belum mendaftarkan IP Perangkat! Silakan klik 'Daftarkan IP Saya' terlebih dahulu.",
		})
		return
	}

	if user.RegisteredIP != clientIP {
		h.Store.SaveAttendanceLog(&models.AttendanceLog{
			Username:       user.Username,
			CheckInTime:    now,
			Type:           req.Type,
			Latitude:       req.Latitude,
			Longitude:      req.Longitude,
			DistanceMeters: services.CalculateDistance(req.Latitude, req.Longitude, services.TargetLat, services.TargetLng),
			IPAddress:      clientIP,
			Status:         "DITOLAK_IP_TIPSEN",
			Notes:          fmt.Sprintf("Percobaan Titip Absen %s! IP Request (%s) != IP Terdaftar (%s)", req.Type, clientIP, user.RegisteredIP),
		})

		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(models.Response{
			Success: false,
			Message: fmt.Sprintf("⚠️ Percobaan Titip Absen Terdeteksi! IP Anda (%s) tidak cocok dengan IP terdaftar akun %s (%s).", clientIP, user.Username, user.RegisteredIP),
		})
		return
	}

	// 3. Check Location (Geofencing <= 800m)
	distance := services.CalculateDistance(req.Latitude, req.Longitude, services.TargetLat, services.TargetLng)
	if distance > services.MaxRadiusMeters {
		h.Store.SaveAttendanceLog(&models.AttendanceLog{
			Username:       user.Username,
			CheckInTime:    now,
			Type:           req.Type,
			Latitude:       req.Latitude,
			Longitude:      req.Longitude,
			DistanceMeters: distance,
			IPAddress:      clientIP,
			Status:         "DITOLAK_LOKASI",
			Notes:          fmt.Sprintf("Di luar radius kantor! Jarak: %.1f meter", distance),
		})

		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(models.Response{
			Success: false,
			Message: fmt.Sprintf("📍 Anda berada di luar radius kantor (Jalan Gatot Subroto I, Denpasar). Jarak: %.1fm", distance),
		})
		return
	}

	// 4. Process Type DATANG vs PULANG
	if req.Type == "DATANG" {
		// Check Cutoff Time (<= 08:30 WITA for Mon-Sat)
		if !services.IsDatangTimeAllowed(now.Hour(), now.Minute()) {
			h.Store.SaveAttendanceLog(&models.AttendanceLog{
				Username:       user.Username,
				CheckInTime:    now,
				Type:           "DATANG",
				Latitude:       req.Latitude,
				Longitude:      req.Longitude,
				DistanceMeters: distance,
				IPAddress:      clientIP,
				Status:         "DITOLAK_WAKTU",
				Notes:          fmt.Sprintf("Terlambat Datang! Absen pada %s (Lebih dari 08:30 WITA)", now.Format("15:04:05")),
			})

			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(models.Response{
				Success: false,
				Message: fmt.Sprintf("🔒 Absensi Datang Ditutup! Waktu saat ini %s WITA (Melebihi batas 08:30 WITA).", now.Format("15:04:05")),
			})
			return
		}

		// Success Datang
		log := &models.AttendanceLog{
			Username:       user.Username,
			CheckInTime:    now,
			Type:           "DATANG",
			Latitude:       req.Latitude,
			Longitude:      req.Longitude,
			DistanceMeters: distance,
			IPAddress:      clientIP,
			Status:         "SUCCESS",
			Notes:          fmt.Sprintf("Berhasil Absen Datang pada %s WITA (%s)", now.Format("15:04:05"), services.GetDayNameIndonesian(weekday)),
		}

		if err := h.Store.SaveAttendanceLog(log); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(models.Response{Success: false, Message: "Gagal menyimpan log presensi"})
			return
		}

		json.NewEncoder(w).Encode(models.Response{
			Success: true,
			Message: fmt.Sprintf("🎉 Presensi DATANG Berhasil! Waktu: %s WITA (%s)", now.Format("15:04:05"), services.GetDayNameIndonesian(weekday)),
			Data:    log,
		})
		return

	} else if req.Type == "PULANG" {
		normalPulangHour := services.GetNormalPulangHour(weekday)
		isEarly := services.IsPulangEarly(weekday, now.Hour(), now.Minute())
		reason := strings.TrimSpace(req.EarlyReason)

		if isEarly && reason == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.Response{
				Success: false,
				Message: fmt.Sprintf("⚠️ Jam pulang hari %s adalah %02d:00 WITA (Waktu saat ini: %s WITA). Wajib menginputkan Alasan Pulang Awal!", services.GetDayNameIndonesian(weekday), normalPulangHour, now.Format("15:04:05")),
			})
			return
		}

		status := "PULANG_NORMAL"
		notes := fmt.Sprintf("Absen Pulang Normal pada %s WITA (%s)", now.Format("15:04:05"), services.GetDayNameIndonesian(weekday))

		if isEarly {
			status = "PULANG_CEPAT"
			notes = fmt.Sprintf("Pulang Awal %s (%02d:00 WITA) pada %s. Alasan: %s", services.GetDayNameIndonesian(weekday), normalPulangHour, now.Format("15:04:05"), reason)
		}

		log := &models.AttendanceLog{
			Username:       user.Username,
			CheckInTime:    now,
			Type:           "PULANG",
			Latitude:       req.Latitude,
			Longitude:      req.Longitude,
			DistanceMeters: distance,
			IPAddress:      clientIP,
			Status:         status,
			EarlyReason:    reason,
			Notes:          notes,
		}

		if err := h.Store.SaveAttendanceLog(log); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(models.Response{Success: false, Message: "Gagal menyimpan log presensi"})
			return
		}

		msg := fmt.Sprintf("🎉 Presensi PULANG Berhasil! Waktu: %s WITA (%s)", now.Format("15:04:05"), services.GetDayNameIndonesian(weekday))
		if isEarly {
			msg = fmt.Sprintf("⚠️ Presensi PULANG AWAL Catat! Waktu: %s WITA | Alasan: %s", now.Format("15:04:05"), reason)
		}

		json.NewEncoder(w).Encode(models.Response{
			Success: true,
			Message: msg,
			Data:    log,
		})
		return

	} else if req.Type == "IZIN" || req.Type == "SAKIT" || req.Type == "LIBUR" {
		reason := strings.TrimSpace(req.EarlyReason)

		// 1. Must be outside office radius (> 800m)
		if distance <= services.MaxRadiusMeters {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.Response{
				Success: false,
				Message: fmt.Sprintf("⚠️ Absen Izin / Sakit hanya berlaku jika Anda berada di LUAR RADIUS KANTOR. Jarak Anda saat ini: %.1fm (Dalam radius kantor).", distance),
			})
			return
		}

		// 2. Reason is mandatory
		if reason == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.Response{
				Success: false,
				Message: "⚠️ Wajib menginputkan Alasan Izin / Sakit!",
			})
			return
		}

		statusName := req.Type
		if statusName == "LIBUR" {
			statusName = "IZIN"
		}

		log := &models.AttendanceLog{
			Username:       user.Username,
			CheckInTime:    now,
			Type:           statusName,
			Latitude:       req.Latitude,
			Longitude:      req.Longitude,
			DistanceMeters: distance,
			IPAddress:      clientIP,
			Status:         statusName,
			EarlyReason:    reason,
			Notes:          fmt.Sprintf("Izin/Sakit pada %s WITA (%s). Alasan: %s", now.Format("15:04:05"), services.GetDayNameIndonesian(weekday), reason),
		}

		if err := h.Store.SaveAttendanceLog(log); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(models.Response{Success: false, Message: "Gagal menyimpan log presensi"})
			return
		}

		json.NewEncoder(w).Encode(models.Response{
			Success: true,
			Message: fmt.Sprintf("🏥 Presensi %s Berhasil Dicatat! Alasan: %s", statusName, reason),
			Data:    log,
		})
		return
	}

	json.NewEncoder(w).Encode(models.Response{Success: false, Message: "Tipe presensi tidak valid (DATANG / PULANG / SAKIT / IZIN)"})
}

func (h *ServerHandler) LogsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	cookie, err := r.Cookie("user_session")
	if err != nil || cookie.Value == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(models.Response{Success: false, Message: "Silakan login terlebih dahulu"})
		return
	}

	logs, err := h.Store.GetLogs(cookie.Value)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.Response{Success: false, Message: "Gagal mengambil log presensi"})
		return
	}

	json.NewEncoder(w).Encode(models.Response{
		Success: true,
		Data:    logs,
	})
}

func (h *ServerHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "user_session",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.Response{Success: true, Message: "Logout berhasil"})
}
