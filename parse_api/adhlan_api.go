package parse_api

import (
	"encoding/json"
	"io"
	"net/http"
)

type PrayerTimesResponse struct {
	Data struct {
		Timings struct {
			Fajr    string `json:"Fajr"`
			Sunrise string `json:"Sunrise"`
			Dhuhr   string `json:"Dhuhr"`
			Asr     string `json:"Asr"`
			Maghrib string `json:"Maghrib"`
			Isha    string `json:"Isha"`
		} `json:"timings"`
	} `json:"data"`
}

func PrayerTimes() (map[string]string, error) {
	url := "https://api.aladhan.com/v1/timings/today?latitude=43.166666&longitude=44.816111&method=14"

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result PrayerTimesResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	times := map[string]string{
		"Fajr":    result.Data.Timings.Fajr,
		"Sunrise": result.Data.Timings.Sunrise,
		"Dhuhr":   result.Data.Timings.Dhuhr,
		"Asr":     result.Data.Timings.Asr,
		"Maghrib": result.Data.Timings.Maghrib,
		"Isha":    result.Data.Timings.Isha,
	}

	return times, nil
}
