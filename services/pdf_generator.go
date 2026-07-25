package services

import (
	"bytes"
	"fmt"
	"time"

	"github.com/go-pdf/fpdf"
	"me.absen/app/models"
	"me.absen/app/static"
)

type DateRow struct {
	No   string
	Date string
	Day  string
}

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

	// Dynamic Date Range: From 15 June 2026 up to today's date (or 30 June 2026 minimum)
	startDate := time.Date(2026, 6, 15, 0, 0, 0, 0, time.FixedZone("WITA", 8*3600))
	endDate := now
	minEndDate := time.Date(2026, 6, 30, 0, 0, 0, 0, time.FixedZone("WITA", 8*3600))
	if endDate.Before(minEndDate) {
		endDate = minEndDate
	}

	var dateRange []DateRow
	noIdx := 1
	for curr := startDate; !curr.After(endDate); curr = curr.AddDate(0, 0, 1) {
		dayStr := GetDayNameIndonesian(curr.Weekday())
		dateStr := curr.Format("02-01-2006")
		dateRange = append(dateRange, DateRow{
			No:   fmt.Sprintf("%d", noIdx),
			Date: dateStr,
			Day:  dayStr,
		})
		noIdx++
	}

	monthTitle := "BULAN: JUNI 2026"
	if now.Month() == time.July {
		monthTitle = "BULAN: JUNI - JULI 2026"
	} else if now.Month() > time.July {
		monthTitle = fmt.Sprintf("BULAN: JUNI - %s %d", stringsToUpper(months[now.Month()]), now.Year())
	}

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// Register & Render Logos (Undiknas & PT Adya Artha Abadi)
	undiknasBytes, err := static.Files.ReadFile("logo_undiknas.png")
	if err == nil {
		pdf.RegisterImageOptionsReader("logo_undiknas", fpdf.ImageOptions{ImageType: "PNG"}, bytes.NewReader(undiknasBytes))
		pdf.ImageOptions("logo_undiknas", 15, 10, 20, 20, false, fpdf.ImageOptions{ImageType: "PNG"}, 0, "")
	}

	companyBytes, err := static.Files.ReadFile("logo_company.png")
	if err == nil {
		pdf.RegisterImageOptionsReader("logo_company", fpdf.ImageOptions{ImageType: "PNG"}, bytes.NewReader(companyBytes))
		pdf.ImageOptions("logo_company", 172, 11, 23, 16, false, fpdf.ImageOptions{ImageType: "PNG"}, 0, "")
	}

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
	pdf.CellFormat(180, 5, monthTitle, "", 1, "L", false, 0, "")
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

		pdf.CellFormat(15, 5.5, item.No, "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 5.5, item.Date, "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 5.5, item.Day, "1", 0, "C", false, 0, "")
		pdf.CellFormat(85, 5.5, statusStr, "1", 1, "C", false, 0, "")
	}

	pdf.Ln(8)

	// Signature Section (No "Mengetahui," line for Kepala Cabang)
	ySig := pdf.GetY()
	// Check page overflow for signature block
	if ySig > 240 {
		pdf.AddPage()
		ySig = 20
	}

	pdf.SetFont("Times", "B", 10.5)
	pdf.SetXY(15, ySig)
	pdf.CellFormat(90, 5, "Kepala Cabang PT. Adya Artha Abadi Bali", "", 0, "C", false, 0, "")
	pdf.SetXY(105, ySig)
	pdf.SetFont("Times", "", 10.5)
	pdf.CellFormat(90, 5, fmt.Sprintf("Denpasar, %s", todayFormatted), "", 1, "C", false, 0, "")

	pdf.SetXY(105, ySig+5)
	pdf.CellFormat(90, 5, "Mahasiswa PKL,", "", 1, "C", false, 0, "")

	pdf.SetXY(15, ySig+28)
	pdf.SetFont("Times", "BU", 10.5)
	pdf.CellFormat(90, 5, "I Made Mas Sugianyar", "", 0, "C", false, 0, "")
	pdf.SetXY(105, ySig+28)
	pdf.CellFormat(90, 5, fullName, "", 1, "C", false, 0, "")

	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("Absensi_PKL_Juni2026_%s.pdf", username)
	return buf.Bytes(), filename, nil
}

func stringsToUpper(s string) string {
	return fmt.Sprintf("%s", bytes.ToUpper([]byte(s)))
}
