package services

import (
	"bytes"
	"fmt"
	"time"

	"github.com/go-pdf/fpdf"
	"me.absen/app/models"
	"me.absen/app/static"
)

type MatrixBlock struct {
	Title    string
	Year     int
	Month    time.Month
	StartDay int
	EndDay   int
}

func getFullStatusAndColor(dateStr string, dayName string, logMap map[string]*models.AttendanceLog, username string) (string, [3]int) {
	if dateStr == "16-06-2026" || dateStr == "17-06-2026" || dateStr == "17-08-2026" {
		return "Libur", [3]int{230, 230, 230} // Light Grey
	}
	if dayName == "Minggu" {
		return "Libur", [3]int{240, 240, 240} // Light Grey
	}
	if log, ok := logMap[dateStr]; ok {
		if log.Type == "SAKIT" || log.Status == "SAKIT" {
			return "Sakit", [3]int{255, 220, 220} // Light Red
		}
		if log.Type == "IZIN" {
			return "Izin", [3]int{255, 245, 200} // Light Yellow
		}
		return "Hadir", [3]int{225, 245, 225} // Light Green
	}
	if username == "deksa" && dateStr == "04-07-2026" {
		return "Sakit", [3]int{255, 220, 220}
	}
	if username == "putra" && dateStr == "14-07-2026" {
		return "Sakit", [3]int{255, 220, 220}
	}
	return "Hadir", [3]int{225, 245, 225}
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

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(10, 10, 10)
	pdf.AddPage()

	// Register & Render Logos (Undiknas & PT Adya Artha Abadi)
	undiknasBytes, err := static.Files.ReadFile("logo_undiknas.png")
	if err == nil {
		pdf.RegisterImageOptionsReader("logo_undiknas", fpdf.ImageOptions{ImageType: "PNG"}, bytes.NewReader(undiknasBytes))
		pdf.ImageOptions("logo_undiknas", 10, 8, 18, 18, false, fpdf.ImageOptions{ImageType: "PNG"}, 0, "")
	}

	companyBytes, err := static.Files.ReadFile("logo_company.png")
	if err == nil {
		pdf.RegisterImageOptionsReader("logo_company", fpdf.ImageOptions{ImageType: "PNG"}, bytes.NewReader(companyBytes))
		pdf.ImageOptions("logo_company", 179, 9, 21, 15, false, fpdf.ImageOptions{ImageType: "PNG"}, 0, "")
	}

	// Header Title
	pdf.SetFont("Times", "B", 12)
	pdf.CellFormat(190, 5, "DAFTAR HADIR PRAKTIK KERJA LAPANGAN (PKL)", "", 1, "C", false, 0, "")
	pdf.SetFont("Times", "B", 10.5)
	pdf.CellFormat(190, 4.5, "UNIVERSITAS PENDIDIKAN NASIONAL (UNDIKNAS) DENPASAR", "", 1, "C", false, 0, "")
	pdf.CellFormat(190, 4.5, "PT ADYA ARTHA ABADI", "", 1, "C", false, 0, "")
	pdf.Ln(4)

	// Student Info Table
	pdf.SetFont("Times", "", 9.5)
	pdf.CellFormat(32, 4.0, "Nama Mahasiswa", "", 0, "L", false, 0, "")
	pdf.CellFormat(4, 4.0, ":", "", 0, "L", false, 0, "")
	pdf.SetFont("Times", "B", 9.5)
	pdf.CellFormat(154, 4.0, fullName, "", 1, "L", false, 0, "")

	pdf.SetFont("Times", "", 9.5)
	pdf.CellFormat(32, 4.0, "NIM", "", 0, "L", false, 0, "")
	pdf.CellFormat(4, 4.0, ":", "", 0, "L", false, 0, "")
	pdf.CellFormat(154, 4.0, nim, "", 1, "L", false, 0, "")

	pdf.CellFormat(32, 4.0, "Program Studi", "", 0, "L", false, 0, "")
	pdf.CellFormat(4, 4.0, ":", "", 0, "L", false, 0, "")
	pdf.CellFormat(154, 4.0, "Manajemen", "", 1, "L", false, 0, "")

	pdf.CellFormat(32, 4.0, "Tempat PKL", "", 0, "L", false, 0, "")
	pdf.CellFormat(4, 4.0, ":", "", 0, "L", false, 0, "")
	pdf.CellFormat(154, 4.0, "PT Adya Artha Abadi", "", 1, "L", false, 0, "")

	pdf.CellFormat(32, 4.0, "Periode PKL", "", 0, "L", false, 0, "")
	pdf.CellFormat(4, 4.0, ":", "", 0, "L", false, 0, "")
	pdf.CellFormat(154, 4.0, "15 Juni s.d. 31 Agustus 2026", "", 1, "L", false, 0, "")

	pdf.Ln(3)

	// 5 Horizontal Matrix Blocks (15-16 days per block for perfect normal unrotated text alignment)
	blocks := []MatrixBlock{
		{"JUNI 2026 (15 - 30 Juni 2026)", 2026, time.June, 15, 30},
		{"JULI 2026 (Bagian 1: 01 - 15 Juli 2026)", 2026, time.July, 1, 15},
		{"JULI 2026 (Bagian 2: 16 - 31 Juli 2026)", 2026, time.July, 16, 31},
		{"AGUSTUS 2026 (Bagian 1: 01 - 15 Agustus 2026)", 2026, time.August, 1, 15},
		{"AGUSTUS 2026 (Bagian 2: 16 - 31 Agustus 2026)", 2026, time.August, 16, 31},
	}

	dayAbbrMap := map[time.Weekday]string{
		time.Sunday:    "Minggu",
		time.Monday:    "Senin",
		time.Tuesday:   "Selasa",
		time.Wednesday: "Rabu",
		time.Thursday:  "Kamis",
		time.Friday:    "Jumat",
		time.Saturday:  "Sabtu",
	}

	lblWidth := 25.0
	availWidth := 190.0 - lblWidth

	for _, b := range blocks {
		pdf.SetFont("Times", "B", 9.0)
		pdf.CellFormat(190, 4.0, fmt.Sprintf("BULAN: %s", b.Title), "", 1, "L", false, 0, "")

		totalDays := b.EndDay - b.StartDay + 1
		cellWidth := availWidth / float64(totalDays)

		// Row 1: Tanggal
		pdf.SetFillColor(235, 235, 235)
		pdf.SetFont("Times", "B", 7.5)
		pdf.CellFormat(lblWidth, 4.2, "Tanggal", "1", 0, "C", true, 0, "")

		for day := b.StartDay; day <= b.EndDay; day++ {
			pdf.CellFormat(cellWidth, 4.2, fmt.Sprintf("%d", day), "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)

		// Row 2: Hari (Full Unrotated Day Names)
		pdf.SetFillColor(245, 245, 245)
		pdf.SetFont("Times", "", 6.2)
		pdf.CellFormat(lblWidth, 4.2, "Hari", "1", 0, "C", true, 0, "")

		for day := b.StartDay; day <= b.EndDay; day++ {
			tDate := time.Date(b.Year, b.Month, day, 0, 0, 0, 0, time.FixedZone("WITA", 8*3600))
			dayName := dayAbbrMap[tDate.Weekday()]
			if tDate.Weekday() == time.Sunday {
				pdf.SetFont("Times", "B", 6.2)
			} else {
				pdf.SetFont("Times", "", 6.2)
			}
			pdf.CellFormat(cellWidth, 4.2, dayName, "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)

		// Row 3: Status Presensi (Full Unrotated Words: Hadir, Libur, Sakit, Izin)
		pdf.SetFont("Times", "B", 7.5)
		pdf.CellFormat(lblWidth, 4.5, "Status Presensi", "1", 0, "C", false, 0, "")

		for day := b.StartDay; day <= b.EndDay; day++ {
			tDate := time.Date(b.Year, b.Month, day, 0, 0, 0, 0, time.FixedZone("WITA", 8*3600))
			dateStr := tDate.Format("02-01-2006")
			dayName := GetDayNameIndonesian(tDate.Weekday())

			statusFull, rgb := getFullStatusAndColor(dateStr, dayName, logMap, username)

			pdf.SetFillColor(rgb[0], rgb[1], rgb[2])
			pdf.SetFont("Times", "B", 6.2)
			if statusFull == "Hadir" {
				pdf.SetFont("Times", "", 6.2)
			}
			pdf.CellFormat(cellWidth, 4.5, statusFull, "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)
		pdf.Ln(2.5)
	}

	pdf.Ln(2)

	// Signature Section (No "Mengetahui,")
	ySig := pdf.GetY()
	if ySig > 250 {
		pdf.AddPage()
		ySig = 20
	}

	pdf.SetFont("Times", "B", 9.5)
	pdf.SetXY(10, ySig)
	pdf.CellFormat(92, 4.5, "Kepala Cabang PT. Adya Artha Abadi Bali", "", 0, "C", false, 0, "")
	pdf.SetXY(108, ySig)
	pdf.SetFont("Times", "", 9.5)
	pdf.CellFormat(92, 4.5, fmt.Sprintf("Denpasar, %s", todayFormatted), "", 1, "C", false, 0, "")

	pdf.SetXY(108, ySig+4.5)
	pdf.CellFormat(92, 4.5, "Mahasiswa PKL,", "", 1, "C", false, 0, "")

	pdf.SetXY(10, ySig+24)
	pdf.SetFont("Times", "BU", 9.5)
	pdf.CellFormat(90, 4.5, "I Made Mas Sugianyar", "", 0, "C", false, 0, "")
	pdf.SetXY(108, ySig+24)
	pdf.CellFormat(92, 4.5, fullName, "", 1, "C", false, 0, "")

	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("Absensi_PKL_JuniAgustus2026_%s.pdf", username)
	return buf.Bytes(), filename, nil
}
