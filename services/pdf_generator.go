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
	pdf.CellFormat(35, 4.2, "Nama Mahasiswa", "", 0, "L", false, 0, "")
	pdf.CellFormat(4, 4.2, ":", "", 0, "L", false, 0, "")
	pdf.SetFont("Times", "B", 9.5)
	pdf.CellFormat(151, 4.2, fullName, "", 1, "L", false, 0, "")

	pdf.SetFont("Times", "", 9.5)
	pdf.CellFormat(35, 4.2, "NIM", "", 0, "L", false, 0, "")
	pdf.CellFormat(4, 4.2, ":", "", 0, "L", false, 0, "")
	pdf.CellFormat(151, 4.2, nim, "", 1, "L", false, 0, "")

	pdf.CellFormat(35, 4.2, "Program Studi", "", 0, "L", false, 0, "")
	pdf.CellFormat(4, 4.2, ":", "", 0, "L", false, 0, "")
	pdf.CellFormat(151, 4.2, "Manajemen", "", 1, "L", false, 0, "")

	pdf.CellFormat(35, 4.2, "Tempat PKL", "", 0, "L", false, 0, "")
	pdf.CellFormat(4, 4.2, ":", "", 0, "L", false, 0, "")
	pdf.CellFormat(151, 4.2, "PT Adya Artha Abadi", "", 1, "L", false, 0, "")

	pdf.CellFormat(35, 4.2, "Periode PKL", "", 0, "L", false, 0, "")
	pdf.CellFormat(4, 4.2, ":", "", 0, "L", false, 0, "")
	pdf.CellFormat(151, 4.2, "15 Juni s.d. 31 Agustus 2026", "", 1, "L", false, 0, "")

	pdf.Ln(3)
	pdf.SetFont("Times", "B", 9.5)
	pdf.CellFormat(190, 4.5, "BULAN: JUNI - AGUSTUS 2026", "", 1, "L", false, 0, "")
	pdf.Ln(2)

	// Side-by-Side 2-Column Table Rendering (Fits ALL 78 Days on 1 Page!)
	half := (len(dateRange) + 1) / 2
	leftRows := dateRange[:half]
	rightRows := dateRange[half:]

	colW := []float64{9, 22, 19, 42}
	gapW := 6.0

	// Table Header (Side by Side)
	pdf.SetFillColor(240, 240, 240)
	pdf.SetFont("Times", "B", 8.5)

	// Left Header
	pdf.CellFormat(colW[0], 5.5, "No", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colW[1], 5.5, "Tanggal", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colW[2], 5.5, "Hari", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colW[3], 5.5, "Hadir / Tidak Hadir", "1", 0, "C", true, 0, "")

	pdf.CellFormat(gapW, 5.5, "", "", 0, "C", false, 0, "")

	// Right Header
	pdf.CellFormat(colW[0], 5.5, "No", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colW[1], 5.5, "Tanggal", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colW[2], 5.5, "Hari", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colW[3], 5.5, "Hadir / Tidak Hadir", "1", 1, "C", true, 0, "")

	// Table Body Rows
	pdf.SetFont("Times", "", 8)
	for i := 0; i < half; i++ {
		itemL := leftRows[i]
		statusL := getStatusStr(itemL, logMap, username)

		if statusL != "Hadir" {
			pdf.SetFont("Times", "B", 8)
		} else {
			pdf.SetFont("Times", "", 8)
		}

		pdf.CellFormat(colW[0], 4.2, itemL.No, "1", 0, "C", false, 0, "")
		pdf.CellFormat(colW[1], 4.2, itemL.Date, "1", 0, "C", false, 0, "")
		pdf.CellFormat(colW[2], 4.2, itemL.Day, "1", 0, "C", false, 0, "")
		pdf.CellFormat(colW[3], 4.2, statusL, "1", 0, "C", false, 0, "")

		pdf.CellFormat(gapW, 4.2, "", "", 0, "C", false, 0, "")

		if i < len(rightRows) {
			itemR := rightRows[i]
			statusR := getStatusStr(itemR, logMap, username)

			if statusR != "Hadir" {
				pdf.SetFont("Times", "B", 8)
			} else {
				pdf.SetFont("Times", "", 8)
			}

			pdf.CellFormat(colW[0], 4.2, itemR.No, "1", 0, "C", false, 0, "")
			pdf.CellFormat(colW[1], 4.2, itemR.Date, "1", 0, "C", false, 0, "")
			pdf.CellFormat(colW[2], 4.2, itemR.Day, "1", 0, "C", false, 0, "")
			pdf.CellFormat(colW[3], 4.2, statusR, "1", 1, "C", false, 0, "")
		} else {
			pdf.CellFormat(colW[0], 4.2, "", "1", 0, "C", false, 0, "")
			pdf.CellFormat(colW[1], 4.2, "", "1", 0, "C", false, 0, "")
			pdf.CellFormat(colW[2], 4.2, "", "1", 0, "C", false, 0, "")
			pdf.CellFormat(colW[3], 4.2, "", "1", 1, "C", false, 0, "")
		}
	}

	pdf.Ln(6)

	// Signature Section
	ySig := pdf.GetY()
	if ySig > 245 {
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
	pdf.CellFormat(92, 4.5, "I Made Mas Sugianyar", "", 0, "C", false, 0, "")
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
