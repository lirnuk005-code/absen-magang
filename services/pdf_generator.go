package services

import (
	"bytes"
	"fmt"
	"time"

	"github.com/go-pdf/fpdf"
	"me.absen/app/models"
)

func GenerateAbsensiPDF(username string, logs []models.AttendanceLog) ([]byte, string, error) {
	fullName := GetFullName(username)
	nim := GetNIM(username)

	now := time.Now().In(time.FixedZone("WITA", 8*3600))
	months := map[time.Month]string{
		time.January:   "Januari",
		time.February:  "Februari",
		time.March:     "Maret",
		time.April:     "April",
		time.May:       "Mei",
		time.June:      "Juni",
		time.July:      "Juli",
		time.August:    "Agustus",
		time.September: "September",
		time.October:   "Oktober",
		time.November:  "November",
		time.December:  "Desember",
	}
	todayFormatted := fmt.Sprintf("%d %s %d", now.Day(), months[now.Month()], now.Year())

	// Build map of logs by DD-MM-YYYY
	logMap := make(map[string]*models.AttendanceLog)
	for _, l := range logs {
		if l.Status != "" && (l.Status == "DITOLAK_DEVICE_TIPSEN" || l.Status == "DITOLAK_IP_TIPSEN") {
			continue
		}
		dateStr := l.CheckInTime.In(time.FixedZone("WITA", 8*3600)).Format("02-01-2006")
		if _, exists := logMap[dateStr]; !exists {
			logMap[dateStr] = &l
		}
	}

	dateRange := []struct {
		No   string
		Date string
		Day  string
	}{
		{"1", "15-06-2026", "Senin"},
		{"2", "16-06-2026", "Selasa"},
		{"3", "17-06-2026", "Rabu"},
		{"4", "18-06-2026", "Kamis"},
		{"5", "19-06-2026", "Jumat"},
		{"6", "20-06-2026", "Sabtu"},
		{"7", "21-06-2026", "Minggu"},
		{"8", "22-06-2026", "Senin"},
		{"9", "23-06-2026", "Selasa"},
		{"10", "24-06-2026", "Rabu"},
		{"11", "25-06-2026", "Kamis"},
		{"12", "26-06-2026", "Jumat"},
		{"13", "27-06-2026", "Sabtu"},
		{"14", "28-06-2026", "Minggu"},
		{"15", "29-06-2026", "Senin"},
		{"16", "30-06-2026", "Selasa"},
	}

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// Header Title
	pdf.SetFont("Times", "B", 13)
	pdf.CellFormat(180, 6, "DAFTAR HADIR PRAKTIK KERJA LAPANGAN (PKL)", "", 1, "C", false, 0, "")
	pdf.SetFont("Times", "B", 11)
	pdf.CellFormat(180, 5, "UNIVERSITAS PENDIDIKAN NASIONAL (UNDIKNAS) DENPASAR", "", 1, "C", false, 0, "")
	pdf.CellFormat(180, 5, "PT ADYA ARTHA ABADI", "", 1, "C", false, 0, "")
	pdf.Ln(6)

	// Student Info Table
	pdf.SetFont("Times", "", 10.5)
	pdf.CellFormat(40, 5, "Nama Mahasiswa", "", 0, "L", false, 0, "")
	pdf.CellFormat(5, 5, ":", "", 0, "L", false, 0, "")
	pdf.SetFont("Times", "B", 10.5)
	pdf.CellFormat(135, 5, fullName, "", 1, "L", false, 0, "")

	pdf.SetFont("Times", "", 10.5)
	pdf.CellFormat(40, 5, "NIM", "", 0, "L", false, 0, "")
	pdf.CellFormat(5, 5, ":", "", 0, "L", false, 0, "")
	pdf.CellFormat(135, 5, nim, "", 1, "L", false, 0, "")

	pdf.CellFormat(40, 5, "Program Studi", "", 0, "L", false, 0, "")
	pdf.CellFormat(5, 5, ":", "", 0, "L", false, 0, "")
	pdf.CellFormat(135, 5, "Manajemen", "", 1, "L", false, 0, "")

	pdf.CellFormat(40, 5, "Tempat PKL", "", 0, "L", false, 0, "")
	pdf.CellFormat(5, 5, ":", "", 0, "L", false, 0, "")
	pdf.CellFormat(135, 5, "PT Adya Artha Abadi", "", 1, "L", false, 0, "")

	pdf.CellFormat(40, 5, "Periode PKL", "", 0, "L", false, 0, "")
	pdf.CellFormat(5, 5, ":", "", 0, "L", false, 0, "")
	pdf.CellFormat(135, 5, "15 s.d. 30 Juni 2026", "", 1, "L", false, 0, "")

	pdf.Ln(4)
	pdf.SetFont("Times", "B", 10.5)
	pdf.CellFormat(180, 5, "BULAN: JUNI 2026", "", 1, "L", false, 0, "")
	pdf.Ln(2)

	// Table Header
	pdf.SetFillColor(240, 240, 240)
	pdf.SetFont("Times", "B", 10)
	pdf.CellFormat(15, 7, "No", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 7, "Tanggal", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 7, "Hari", "1", 0, "C", true, 0, "")
	pdf.CellFormat(85, 7, "Hadir / Tidak Hadir", "1", 1, "C", true, 0, "")

	// Table Body Rows
	pdf.SetFont("Times", "", 9.5)
	for _, item := range dateRange {
		statusStr := "Hadir"
		if item.Date == "16-06-2026" || item.Date == "17-06-2026" {
			statusStr = "Libur Hari Raya Galungan"
		} else if item.Day == "Minggu" {
			statusStr = "Libur Hari Minggu"
		} else if log, ok := logMap[item.Date]; ok {
			if log.Type == "SAKIT" || log.Status == "SAKIT" {
				statusStr = "Sakit"
			} else if log.Type == "IZIN" {
				statusStr = "Izin"
			}
		} else if username == "deksa" && item.Date == "04-07-2026" {
			statusStr = "Sakit"
		} else if username == "putra" && item.Date == "14-07-2026" {
			statusStr = "Sakit"
		}

		if statusStr != "Hadir" {
			pdf.SetFont("Times", "B", 9.5)
		} else {
			pdf.SetFont("Times", "", 9.5)
		}

		pdf.CellFormat(15, 6, item.No, "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 6, item.Date, "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 6, item.Day, "1", 0, "C", false, 0, "")
		pdf.CellFormat(85, 6, statusStr, "1", 1, "C", false, 0, "")
	}

	pdf.Ln(10)

	// Signature Section
	ySig := pdf.GetY()
	pdf.SetFont("Times", "", 10.5)
	pdf.SetXY(15, ySig)
	pdf.CellFormat(90, 5, "Mengetahui,", "", 0, "C", false, 0, "")
	pdf.SetXY(105, ySig)
	pdf.CellFormat(90, 5, fmt.Sprintf("Denpasar, %s", todayFormatted), "", 1, "C", false, 0, "")

	pdf.SetXY(15, ySig+5)
	pdf.SetFont("Times", "B", 10.5)
	pdf.CellFormat(90, 5, "Kepala Cabang PT. Adya Artha Abadi Bali", "", 0, "C", false, 0, "")
	pdf.SetXY(105, ySig+5)
	pdf.SetFont("Times", "", 10.5)
	pdf.CellFormat(90, 5, "Mahasiswa PKL,", "", 1, "C", false, 0, "")

	pdf.SetXY(15, ySig+28)
	pdf.SetFont("Times", "BU", 10.5)
	pdf.CellFormat(90, 5, "I Made Mas Sugianyar", "", 0, "C", false, 0, "")
	pdf.SetXY(105, ySig+28)
	pdf.CellFormat(90, 5, fullName, "", 1, "C", false, 0, "")

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("Absensi_PKL_Juni2026_%s.pdf", username)
	return buf.Bytes(), filename, nil
}
