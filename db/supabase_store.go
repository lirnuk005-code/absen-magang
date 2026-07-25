package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"me.absen/app/models"
)

type SupabaseStore struct {
	URL        string
	ServiceKey string
	Client     *http.Client
}

func NewSupabaseStore(url, serviceKey string) *SupabaseStore {
	return &SupabaseStore{
		URL:        url,
		ServiceKey: serviceKey,
		Client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *SupabaseStore) GetUser(username string) (*models.User, error) {
	reqURL := fmt.Sprintf("%s/rest/v1/profiles?username=eq.%s", s.URL, username)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("apikey", s.ServiceKey)
	req.Header.Set("Authorization", "Bearer "+s.ServiceKey)

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var users []models.User
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &users); err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return &users[0], nil
}

func (s *SupabaseStore) RegisterIP(username, ip string) error {
	// First check if user already has a registered IP
	user, err := s.GetUser(username)
	if err != nil {
		return err
	}

	if user.RegisteredIP != "" {
		return fmt.Errorf("Akun ini sudah terdaftar dengan IP %s. Setiap akun hanya diizinkan mendaftarkan 1 IP Address", user.RegisteredIP)
	}

	reqURL := fmt.Sprintf("%s/rest/v1/profiles?username=eq.%s", s.URL, username)
	payload := map[string]string{"registered_ip": ip}
	jsonBytes, _ := json.Marshal(payload)

	req, err := http.NewRequest("PATCH", reqURL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}

	req.Header.Set("apikey", s.ServiceKey)
	req.Header.Set("Authorization", "Bearer "+s.ServiceKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	resp, err := s.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("gagal update IP di Supabase: %s", string(body))
	}

	// 2. Update/Sync all historical attendance logs for this user with the newly registered IP
	logsReqURL := fmt.Sprintf("%s/rest/v1/attendance_logs?username=eq.%s", s.URL, username)
	logsPayload := map[string]string{"ip_address": ip}
	logsBytes, _ := json.Marshal(logsPayload)

	logsReq, errLogs := http.NewRequest("PATCH", logsReqURL, bytes.NewBuffer(logsBytes))
	if errLogs == nil {
		logsReq.Header.Set("apikey", s.ServiceKey)
		logsReq.Header.Set("Authorization", "Bearer "+s.ServiceKey)
		logsReq.Header.Set("Content-Type", "application/json")

		respLogs, errDo := s.Client.Do(logsReq)
		if errDo == nil {
			respLogs.Body.Close()
		}
	}

	return nil
}

func (s *SupabaseStore) SaveAttendanceLog(log *models.AttendanceLog) error {
	reqURL := fmt.Sprintf("%s/rest/v1/attendance_logs", s.URL)
	jsonBytes, _ := json.Marshal(log)

	req, err := http.NewRequest("POST", reqURL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}

	req.Header.Set("apikey", s.ServiceKey)
	req.Header.Set("Authorization", "Bearer "+s.ServiceKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=minimal")

	resp, err := s.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("gagal simpan log di Supabase: %s", string(body))
	}

	return nil
}

// GetLogs returns attendance logs ONLY for the specified username
func (s *SupabaseStore) GetLogs(username string) ([]models.AttendanceLog, error) {
	reqURL := fmt.Sprintf("%s/rest/v1/attendance_logs?username=eq.%s&order=check_in_time.desc&limit=1000", s.URL, username)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("apikey", s.ServiceKey)
	req.Header.Set("Authorization", "Bearer "+s.ServiceKey)

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var logs []models.AttendanceLog
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &logs); err != nil {
		return nil, err
	}

	return logs, nil
}
