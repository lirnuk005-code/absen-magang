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

func getStatusStr(item DateRow, logMap map[string]*models.AttendanceLog, username string) string {
	if item.Date == "16-06-2026" || item.Date == "17-06-2026" {
		return "Libur Hari Raya Galungan"
	}
	if item.Date == "17-08-2026" {
		return "Libur Kemerdekaan RI"
	}
	if item.Day == "Minggu" {
		return "Libur Hari Minggu"
	}
	if log, ok := logMap[item.Date]; ok {
		if log.Type == "SAKIT" || log.Status == "SAKIT" {
			return "Sakit"
		}
		if log.Type == "IZIN" {
			return "Izin"
		}
		return "Hadir"
	}
	if username == "deksa" && item.Date == "04-07-2026" {
		return "Sakit"
	}
	if username == "putra" && item.Date == "14-07-2026" {
		return "Sakit"
	}
	return "Hadir"
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

	// Full Internship Date Range: 15 Juni 2026 s.d. 31 Agustus 2026 (78 Days total)
	startDate := time.Date(2026, 6, 15, 0, 0, 0, 0, time.FixedZone("WITA", 8*3600))
	endDate := time.Date(2026, 8, 31, 0, 0, 0, 0, time.FixedZone("WITA", 8*3600))

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

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
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
	pdf.CellFormat(135, 5, "15 Juni s.d. 31 Agustus 2026", "", 1, "L", false, 0, "")

	pdf.Ln(4)
	pdf.SetFont("Times", "B", 10.5)
	pdf.CellFormat(180, 5, "BULAN: JUNI - AGUSTUS 2026", "", 1, "L", false, 0, "")
	pdf.Ln(2)

	// Single Column Continuous Table
	colW := []float64{15, 35, 35, 95}

	pdf.SetFillColor(240, 240, 240)
	pdf.SetFont("Times", "B", 10)
	pdf.CellFormat(colW[0], 6.5, "No", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colW[1], 6.5, "Tanggal", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colW[2], 6.5, "Hari", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colW[3], 6.5, "Hadir / Tidak Hadir", "1", 1, "C", true, 0, "")

	pdf.SetFont("Times", "", 9.5)
	for _, item := range dateRange {
		statusStr := getStatusStr(item, logMap, username)

		if statusStr != "Hadir" {
			pdf.SetFont("Times", "B", 9.5)
		} else {
			pdf.SetFont("Times", "", 9.5)
		}

		pdf.CellFormat(colW[0], 5.2, item.No, "1", 0, "C", false, 0, "")
		pdf.CellFormat(colW[1], 5.2, item.Date, "1", 0, "C", false, 0, "")
		pdf.CellFormat(colW[2], 5.2, item.Day, "1", 0, "C", false, 0, "")
		pdf.CellFormat(colW[3], 5.2, statusStr, "1", 1, "C", false, 0, "")
	}

	pdf.Ln(8)

	// Signature Section (No "Mengetahui,")
	ySig := pdf.GetY()
	if ySig > 245 {
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

	filename := fmt.Sprintf("Absensi_PKL_JuniAgustus2026_%s.pdf", username)
	return buf.Bytes(), filename, nil
}
