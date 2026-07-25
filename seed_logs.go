package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	SupabaseURL = "https://sqlxflkwmyogxmhdtlgu.supabase.co"
	ServiceKey  = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InNxbHhmbGt3bXlvZ3htaGR0bGd1Iiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImlhdCI6MTc4NDkxODE4NSwiZXhwIjoyMTAwNDk0MTg1fQ.7YkTWWijC1VjnH9pt4hlyCVNvKAmk7nBKy1x98tacQ0"
)

type AttendanceLogSeed struct {
	Username       string    `json:"username"`
	CheckInTime    time.Time `json:"check_in_time"`
	Type           string    `json:"type"`
	Latitude       float64   `json:"latitude"`
	Longitude      float64   `json:"longitude"`
	DistanceMeters float64   `json:"distance_meters"`
	IPAddress      string    `json:"ip_address"`
	Status         string    `json:"status"`
	EarlyReason    string    `json:"early_reason"`
	Notes          string    `json:"notes"`
}

func main() {
	client := &http.Client{Timeout: 30 * time.Second}

	// 1. Delete all existing logs from Supabase
	delReq, _ := http.NewRequest("DELETE", SupabaseURL+"/rest/v1/attendance_logs?id=neq.00000000-0000-0000-0000-000000000000", nil)
	delReq.Header.Set("apikey", ServiceKey)
	delReq.Header.Set("Authorization", "Bearer "+ServiceKey)
	respDel, err := client.Do(delReq)
	if err != nil {
		fmt.Printf("Error deleting logs: %v\n", err)
	} else {
		respDel.Body.Close()
		fmt.Println("Cleared old attendance logs.")
	}

	// 2. Register user IPs
	userIPs := map[string]string{
		"chris": "180.252.10.1",
		"deksa": "180.252.10.2",
		"putra": "180.252.10.3",
	}

	for u, ip := range userIPs {
		patchPayload, _ := json.Marshal(map[string]string{"registered_ip": ip})
		patchReq, _ := http.NewRequest("PATCH", SupabaseURL+"/rest/v1/profiles?username=eq."+u, bytes.NewBuffer(patchPayload))
		patchReq.Header.Set("apikey", ServiceKey)
		patchReq.Header.Set("Authorization", "Bearer "+ServiceKey)
		patchReq.Header.Set("Content-Type", "application/json")
		r, e := client.Do(patchReq)
		if e == nil {
			r.Body.Close()
		}
	}

	// 3. Generate logs from 15 June 2026 to 24 July 2026
	users := []string{"chris", "deksa", "putra"}
	var allLogs []AttendanceLogSeed

	wita := time.FixedZone("WITA", 8*3600)

	// Iterate day by day explicitly
	current := time.Date(2026, 6, 15, 0, 0, 0, 0, wita)
	end := time.Date(2026, 7, 24, 0, 0, 0, 0, wita)

	for !current.After(end) {
		year, month, day := current.Date()
		weekday := current.Weekday()

		fmt.Printf("Generating day: %02d-%02d-%d (%s)\n", day, month, year, weekday)

		for uIdx, user := range users {
			ip := userIPs[user]

			// Case 1: Hari Raya Galungan (16 & 17 Juni 2026)
			if month == 6 && (day == 16 || day == 17) {
				t := time.Date(year, month, day, 8, 0, 0, 0, wita)
				allLogs = append(allLogs, AttendanceLogSeed{
					Username:       user,
					CheckInTime:    t,
					Type:           "LIBUR",
					Latitude:       -8.6366,
					Longitude:      115.2223,
					DistanceMeters: 0,
					IPAddress:      ip,
					Status:         "LIBUR_GALUNGAN",
					Notes:          "Libur Hari Raya Galungan",
				})
				continue
			}

			// Case 2: Hari Minggu
			if weekday == time.Sunday {
				t := time.Date(year, month, day, 8, 0, 0, 0, wita)
				allLogs = append(allLogs, AttendanceLogSeed{
					Username:       user,
					CheckInTime:    t,
					Type:           "LIBUR",
					Latitude:       -8.6366,
					Longitude:      115.2223,
					DistanceMeters: 0,
					IPAddress:      ip,
					Status:         "LIBUR_HARI_MINGGU",
					Notes:          "Libur Hari Minggu (Sistem Otomatis)",
				})
				continue
			}

			// Case 3: Sakit Deksa (4 Juli 2026 - Sabtu)
			if month == 7 && day == 4 && user == "deksa" {
				t := time.Date(year, month, day, 8, 0, 0, 0, wita)
				allLogs = append(allLogs, AttendanceLogSeed{
					Username:       user,
					CheckInTime:    t,
					Type:           "SAKIT",
					Latitude:       -8.6366,
					Longitude:      115.2223,
					DistanceMeters: 0,
					IPAddress:      ip,
					Status:         "SAKIT",
					EarlyReason:    "Libur Karena Sakit",
					Notes:          "Izin Sakit (Libur Karena Sakit)",
				})
				continue
			}

			// Case 4: Sakit Putra (14 Juli 2026 - Selasa)
			if month == 7 && day == 14 && user == "putra" {
				t := time.Date(year, month, day, 8, 0, 0, 0, wita)
				allLogs = append(allLogs, AttendanceLogSeed{
					Username:       user,
					CheckInTime:    t,
					Type:           "SAKIT",
					Latitude:       -8.6366,
					Longitude:      115.2223,
					DistanceMeters: 0,
					IPAddress:      ip,
					Status:         "SAKIT",
					EarlyReason:    "Libur Karena Sakit",
					Notes:          "Izin Sakit (Libur Karena Sakit)",
				})
				continue
			}

			// Case 5: Normal Work Day
			datangMin := 5 + (uIdx * 3) + (day % 4)
			pulangMin := 5 + (uIdx * 2) + (day % 3)

			tDatang := time.Date(year, month, day, 8, datangMin, 0, 0, wita)
			allLogs = append(allLogs, AttendanceLogSeed{
				Username:       user,
				CheckInTime:    tDatang,
				Type:           "DATANG",
				Latitude:       -8.6366,
				Longitude:      115.2223,
				DistanceMeters: 0,
				IPAddress:      ip,
				Status:         "SUCCESS",
				Notes:          fmt.Sprintf("Berhasil Absen Datang pada %02d:%02d WITA", 8, datangMin),
			})

			pulangHour := 16
			if weekday == time.Saturday {
				pulangHour = 13
			}

			tPulang := time.Date(year, month, day, pulangHour, pulangMin, 0, 0, wita)
			allLogs = append(allLogs, AttendanceLogSeed{
				Username:       user,
				CheckInTime:    tPulang,
				Type:           "PULANG",
				Latitude:       -8.6366,
				Longitude:      115.2223,
				DistanceMeters: 0,
				IPAddress:      ip,
				Status:         "PULANG_NORMAL",
				Notes:          fmt.Sprintf("Absen Pulang Normal pada %02d:%02d WITA", pulangHour, pulangMin),
			})
		}

		current = current.AddDate(0, 0, 1)
	}

	fmt.Printf("Total logs generated: %d\n", len(allLogs))

	// 4. Insert one-by-one or small chunks (20 logs per batch) with error logging
	batchSize := 25
	for i := 0; i < len(allLogs); i += batchSize {
		endIdx := i + batchSize
		if endIdx > len(allLogs) {
			endIdx = len(allLogs)
		}
		batch := allLogs[i:endIdx]

		jsonBytes, _ := json.Marshal(batch)
		postReq, _ := http.NewRequest("POST", SupabaseURL+"/rest/v1/attendance_logs", bytes.NewBuffer(jsonBytes))
		postReq.Header.Set("apikey", ServiceKey)
		postReq.Header.Set("Authorization", "Bearer "+ServiceKey)
		postReq.Header.Set("Content-Type", "application/json")
		postReq.Header.Set("Prefer", "return=minimal")

		r, err := client.Do(postReq)
		if err != nil {
			fmt.Printf("Error posting batch %d-%d: %v\n", i, endIdx, err)
		} else {
			body, _ := io.ReadAll(r.Body)
			if r.StatusCode >= 400 {
				fmt.Printf("Batch %d-%d failed (%d): %s\n", i, endIdx, r.StatusCode, string(body))
			} else {
				fmt.Printf("Batch %d-%d inserted successfully!\n", i, endIdx)
			}
			r.Body.Close()
		}
		time.Sleep(100 * time.Millisecond) // slight delay to ensure clean DB commits
	}

	fmt.Println("Complete! All attendance records from 15 June 2026 to 24 July 2026 seeded cleanly.")
}
