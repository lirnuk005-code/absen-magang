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

func getRealStatusStr(item DateRow, logMap map[string]*models.AttendanceLog) string {
	if item.Date == "16-06-2026" || item.Date == "17-06-2026" {
		return "Libur Hari Raya Galungan"
	}
	if item.Date == "17-08-2026" {
		return "Libur Hari Kemerdekaan RI"
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
		if log.Type == "LIBUR" {
			return "Libur"
		}
		return "Hadir"
	}
	return "Tidak Hadir"
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

	// Dynamic Internship Date Range: 15 Juni 2026 up to TODAY (or max 31 Agustus 2026)
	startDate := time.Date(2026, 6, 15, 0, 0, 0, 0, time.FixedZone("WITA", 8*3600))
	endDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.FixedZone("WITA", 8*3600))
	maxDate := time.Date(2026, 8, 31, 0, 0, 0, 0, time.FixedZone("WITA", 8*3600))

	if endDate.Before(startDate) {
		endDate = startDate
	}
	if endDate.After(maxDate) {
		endDate = maxDate
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

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(12, 5, 12)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()

	// 1. Register & Render Logos
	undiknasBytes, err := static.Files.ReadFile("logo_undiknas.png")
	if err == nil {
		pdf.RegisterImageOptionsReader("logo_undiknas", fpdf.ImageOptions{ImageType: "PNG"}, bytes.NewReader(undiknasBytes))
		pdf.ImageOptions("logo_undiknas", 12, 5, 15, 15, false, fpdf.ImageOptions{ImageType: "PNG"}, 0, "")
	}

	companyBytes, err := static.Files.ReadFile("logo_company.png")
	if err == nil {
		pdf.RegisterImageOptionsReader("logo_company", fpdf.ImageOptions{ImageType: "PNG"}, bytes.NewReader(companyBytes))
		pdf.ImageOptions("logo_company", 180, 6.5, 18, 12, false, fpdf.ImageOptions{ImageType: "PNG"}, 0, "")
	}

	// 2. Header Title Text
	pdf.SetY(5)
	pdf.SetFont("Times", "B", 11)
	pdf.CellFormat(186, 3.8, "DAFTAR HADIR PRAKTIK KERJA LAPANGAN (PKL)", "", 1, "C", false, 0, "")
	pdf.SetFont("Times", "B", 9.5)
	pdf.CellFormat(186, 3.5, "UNIVERSITAS PENDIDIKAN NASIONAL (UNDIKNAS) DENPASAR", "", 1, "C", false, 0, "")
	pdf.CellFormat(186, 3.5, "PT ADYA ARTHA ABADI", "", 1, "C", false, 0, "")
	pdf.SetFont("Times", "I", 7.8)
	pdf.CellFormat(186, 3.0, "Alamat: Jalan Gatot Subroto I, Tonja, Denpasar Utara, Denpasar, Bali", "", 1, "C", false, 0, "")

	// 3. Render Official Kop Double Line
	yLine := 21.2
	pdf.SetLineWidth(0.7)
	pdf.Line(12, yLine, 198, yLine)
	pdf.SetLineWidth(0.2)
	pdf.Line(12, yLine+0.7, 198, yLine+0.7)
	pdf.SetLineWidth(0.2)

	// 4. Student Info Metadata Box
	pdf.SetY(23.8)
	pdf.SetFont("Times", "", 8.5)
	pdf.CellFormat(28, 3.2, "Nama Mahasiswa", "", 0, "L", false, 0, "")
	pdf.CellFormat(3, 3.2, ":", "", 0, "L", false, 0, "")
	pdf.SetFont("Times", "B", 8.5)
	pdf.CellFormat(60, 3.2, fullName, "", 0, "L", false, 0, "")

	pdf.SetFont("Times", "", 8.5)
	pdf.CellFormat(25, 3.2, "Tempat PKL", "", 0, "L", false, 0, "")
	pdf.CellFormat(3, 3.2, ":", "", 0, "L", false, 0, "")
	pdf.CellFormat(67, 3.2, "PT Adya Artha Abadi", "", 1, "L", false, 0, "")

	pdf.CellFormat(28, 3.2, "NIM / Prodi", "", 0, "L", false, 0, "")
	pdf.CellFormat(3, 3.2, ":", "", 0, "L", false, 0, "")
	pdf.CellFormat(60, 3.2, fmt.Sprintf("%s / Manajemen", nim), "", 0, "L", false, 0, "")

	pdf.CellFormat(25, 3.2, "Periode PKL", "", 0, "L", false, 0, "")
	pdf.CellFormat(3, 3.2, ":", "", 0, "L", false, 0, "")
	pdf.CellFormat(67, 3.2, "15 Juni s.d. 31 Agustus 2026", "", 1, "L", false, 0, "")

	pdf.Ln(2.5)

	// 5. Dynamic Row Height Calculation for Perfect Spacing
	totalRows := len(dateRange)
	rowH := 2.80
	if totalRows <= 45 {
		rowH = 3.80 // Generous row height when days <= 45 (up to late July)
	} else if totalRows <= 60 {
		rowH = 3.20
	}

	colW := []float64{14, 40, 40, 92}

	pdf.SetFillColor(230, 235, 245)
	pdf.SetFont("Times", "B", 8.5)
	pdf.CellFormat(colW[0], 4.2, "No", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colW[1], 4.2, "Tanggal", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colW[2], 4.2, "Hari", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colW[3], 4.2, "Status Presensi", "1", 1, "C", true, 0, "")

	pdf.SetFont("Times", "", 8)
	for i, item := range dateRange {
		statusStr := getRealStatusStr(item, logMap)

		fill := false
		if i%2 == 1 {
			pdf.SetFillColor(248, 249, 250)
			fill = true
		}

		if statusStr != "Hadir" {
			pdf.SetFont("Times", "B", 8)
		} else {
			pdf.SetFont("Times", "", 8)
		}

		pdf.CellFormat(colW[0], rowH, item.No, "1", 0, "C", fill, 0, "")
		pdf.CellFormat(colW[1], rowH, item.Date, "1", 0, "C", fill, 0, "")
		pdf.CellFormat(colW[2], rowH, item.Day, "1", 0, "C", fill, 0, "")
		pdf.CellFormat(colW[3], rowH, statusStr, "1", 1, "C", fill, 0, "")
	}

	// 6. Generous Comfort Gap before Signature Section (Min 8.0mm space!)
	pdf.Ln(8.0)

	ySig := pdf.GetY()
	if ySig > 255 {
		ySig = 255
	}

	pdf.SetFont("Times", "", 8.5)
	pdf.SetXY(12, ySig)
	pdf.CellFormat(90, 3.5, fmt.Sprintf("Denpasar, %s", todayFormatted), "", 0, "C", false, 0, "")
	pdf.SetXY(108, ySig)
	pdf.CellFormat(94, 3.5, fmt.Sprintf("Denpasar, %s", todayFormatted), "", 1, "C", false, 0, "")

	pdf.SetFont("Times", "B", 8.5)
	pdf.SetXY(12, ySig+3.5)
	pdf.CellFormat(90, 3.5, "Kepala Cabang PT. Adya Artha Abadi Bali,", "", 0, "C", false, 0, "")
	pdf.SetXY(108, ySig+3.5)
	pdf.SetFont("Times", "", 8.5)
	pdf.CellFormat(94, 3.5, "Mahasiswa PKL,", "", 1, "C", false, 0, "")

	pdf.SetXY(12, ySig+18)
	pdf.SetFont("Times", "BU", 8.5)
	pdf.CellFormat(90, 3.5, "I Made Mas Sugianyar", "", 0, "C", false, 0, "")
	pdf.SetXY(108, ySig+18)
	pdf.CellFormat(94, 3.5, fullName, "", 1, "C", false, 0, "")

	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("Absensi_PKL_JuniAgustus2026_%s.pdf", username)
	return buf.Bytes(), filename, nil
}
