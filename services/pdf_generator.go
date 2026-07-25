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

func getFullStatusStr(item DateRow, logMap map[string]*models.AttendanceLog, username string) string {
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

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 12, 15)
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
	pdf.SetFont("Times", "", 10)
	pdf.CellFormat(36, 4.5, "Nama Mahasiswa", "", 0, "L", false, 0, "")
	pdf.CellFormat(4, 4.5, ":", "", 0, "L", false, 0, "")
	pdf.SetFont("Times", "B", 10)
	pdf.CellFormat(140, 4.5, fullName, "", 1, "L", false, 0, "")

	pdf.SetFont("Times", "", 10)
	pdf.CellFormat(36, 4.5, "NIM", "", 0, "L", false, 0, "")
	pdf.CellFormat(4, 4.5, ":", "", 0, "L", false, 0, "")
	pdf.CellFormat(140, 4.5, nim, "", 1, "L", false, 0, "")

	pdf.CellFormat(36, 4.5, "Program Studi", "", 0, "L", false, 0, "")
	pdf.CellFormat(4, 4.5, ":", "", 0, "L", false, 0, "")
	pdf.CellFormat(140, 4.5, "Manajemen", "", 1, "L", false, 0, "")

	pdf.CellFormat(36, 4.5, "Tempat PKL", "", 0, "L", false, 0, "")
	pdf.CellFormat(4, 4.5, ":", "", 0, "L", false, 0, "")
	pdf.CellFormat(140, 4.5, "PT Adya Artha Abadi", "", 1, "L", false, 0, "")

	pdf.CellFormat(36, 4.5, "Periode PKL", "", 0, "L", false, 0, "")
	pdf.CellFormat(4, 4.5, ":", "", 0, "L", false, 0, "")
	pdf.CellFormat(140, 4.5, "15 Juni s.d. 31 Agustus 2026", "", 1, "L", false, 0, "")

	pdf.Ln(4)

	// 3 Monthly Sections: Juni 2026, Juli 2026, Agustus 2026
	sections := []struct {
		Title    string
		Year     int
		Month    time.Month
		StartDay int
		EndDay   int
	}{
		{"BULAN: JUNI 2026 (15 - 30 JUNI 2026)", 2026, time.June, 15, 30},
		{"BULAN: JULI 2026 (01 - 31 JULI 2026)", 2026, time.July, 1, 31},
		{"BULAN: AGUSTUS 2026 (01 - 31 AGUSTUS 2026)", 2026, time.August, 1, 31},
	}

	colW := []float64{15, 35, 35, 95}
	globalNo := 1

	for idx, sec := range sections {
		// Clean page break before August for perfect 2-page document layout
		if idx == 2 && pdf.GetY() > 170 {
			pdf.AddPage()
		}

		// Section Header
		pdf.SetFont("Times", "B", 10.5)
		pdf.CellFormat(180, 6, sec.Title, "", 1, "L", false, 0, "")

		// Table Header
		pdf.SetFillColor(240, 240, 240)
		pdf.SetFont("Times", "B", 9.5)
		pdf.CellFormat(colW[0], 5.5, "No", "1", 0, "C", true, 0, "")
		pdf.CellFormat(colW[1], 5.5, "Tanggal", "1", 0, "C", true, 0, "")
		pdf.CellFormat(colW[2], 5.5, "Hari", "1", 0, "C", true, 0, "")
		pdf.CellFormat(colW[3], 5.5, "Hadir / Tidak Hadir", "1", 1, "C", true, 0, "")

		pdf.SetFont("Times", "", 9)
		for day := sec.StartDay; day <= sec.EndDay; day++ {
			tDate := time.Date(sec.Year, sec.Month, day, 0, 0, 0, 0, time.FixedZone("WITA", 8*3600))
			dateStr := tDate.Format("02-01-2006")
			dayName := GetDayNameIndonesian(tDate.Weekday())

			item := DateRow{No: fmt.Sprintf("%d", globalNo), Date: dateStr, Day: dayName}
			statusStr := getFullStatusStr(item, logMap, username)

			if statusStr != "Hadir" {
				pdf.SetFont("Times", "B", 9)
			} else {
				pdf.SetFont("Times", "", 9)
			}

			pdf.CellFormat(colW[0], 4.6, item.No, "1", 0, "C", false, 0, "")
			pdf.CellFormat(colW[1], 4.6, item.Date, "1", 0, "C", false, 0, "")
			pdf.CellFormat(colW[2], 4.6, item.Day, "1", 0, "C", false, 0, "")
			pdf.CellFormat(colW[3], 4.6, statusStr, "1", 1, "C", false, 0, "")

			globalNo++
		}

		pdf.Ln(4)
	}

	pdf.Ln(4)

	// Signature Section (No "Mengetahui,")
	ySig := pdf.GetY()
	if ySig > 235 {
		pdf.AddPage()
		ySig = 20
	}

	pdf.SetFont("Times", "B", 10)
	pdf.SetXY(15, ySig)
	pdf.CellFormat(90, 5, "Kepala Cabang PT. Adya Artha Abadi Bali", "", 0, "C", false, 0, "")
	pdf.SetXY(105, ySig)
	pdf.SetFont("Times", "", 10)
	pdf.CellFormat(90, 5, fmt.Sprintf("Denpasar, %s", todayFormatted), "", 1, "C", false, 0, "")

	pdf.SetXY(105, ySig+5)
	pdf.CellFormat(90, 5, "Mahasiswa PKL,", "", 1, "C", false, 0, "")

	pdf.SetXY(15, ySig+26)
	pdf.SetFont("Times", "BU", 10)
	pdf.CellFormat(90, 5, "I Made Mas Sugianyar", "", 0, "C", false, 0, "")
	pdf.SetXY(105, ySig+26)
	pdf.CellFormat(90, 5, fullName, "", 1, "C", false, 0, "")

	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("Absensi_PKL_JuniAgustus2026_%s.pdf", username)
	return buf.Bytes(), filename, nil
}
