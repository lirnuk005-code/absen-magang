package services

import (
	"math"
	"time"
)

const (
	// Target location: Jalan Gatot Subroto I, Tonja, Denpasar Utara, Denpasar, Bali
	TargetLat       = -8.6366
	TargetLng       = 115.2223
	MaxRadiusMeters = 800.0 // 800m GPS tolerance

	// Jam Kerja Datang (Senin - Sabtu): 08:00 (Cutoff 08:30)
	CutoffDatangHour   = 8
	CutoffDatangMinute = 30

	// Jam Kerja Pulang:
	// Senin - Jumat : 16:00 WITA
	// Sabtu         : 13:00 WITA
	PulangWeekdayHour  = 16
	PulangSaturdayHour = 13
)

// CalculateDistance computes Haversine distance in meters
func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371000 // meters

	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLon := (lon2 - lon1) * (math.Pi / 180.0)

	rLat1 := lat1 * (math.Pi / 180.0)
	rLat2 := lat2 * (math.Pi / 180.0)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(rLat1)*math.Cos(rLat2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

// IsDatangTimeAllowed checks if time is <= 08:30 WITA
func IsDatangTimeAllowed(hour, minute int) bool {
	if hour < CutoffDatangHour {
		return true
	}
	if hour == CutoffDatangHour && minute <= CutoffDatangMinute {
		return true
	}
	return false
}

// GetNormalPulangHour returns 16 for Mon-Fri, and 13 for Saturday
func GetNormalPulangHour(weekday time.Weekday) int {
	if weekday == time.Saturday {
		return PulangSaturdayHour
	}
	return PulangWeekdayHour
}

// IsPulangEarly checks if current hour is before normal check-out hour for that day
func IsPulangEarly(weekday time.Weekday, hour, minute int) bool {
	normalHour := GetNormalPulangHour(weekday)
	if hour < normalHour {
		return true
	}
	return false
}

// GetDayNameIndonesian returns day name in Indonesian
func GetDayNameIndonesian(weekday time.Weekday) string {
	days := map[time.Weekday]string{
		time.Sunday:    "Minggu",
		time.Monday:    "Senin",
		time.Tuesday:   "Selasa",
		time.Wednesday: "Rabu",
		time.Thursday:  "Kamis",
		time.Friday:    "Jumat",
		time.Saturday:  "Sabtu",
	}
	return days[weekday]
}

// GetFullName returns student's official full name
func GetFullName(username string) string {
	switch username {
	case "chris":
		return "Christopher Natanael"
	case "deksa":
		return "I Kadek Mahesa Parwata Gandhi"
	case "putra":
		return "I Gede Agus Nova Pratama Putra"
	default:
		return username
	}
}

// GetNIM returns student's official NIM
func GetNIM(username string) string {
	switch username {
	case "chris":
		return "12310299"
	case "deksa":
		return "12310300"
	case "putra":
		return "12310311"
	default:
		return "-"
	}
}
