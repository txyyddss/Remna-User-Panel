package httpapi

import (
	"net/http"
	"strings"
	"time"
)

func (s *Server) statisticsWindow(w http.ResponseWriter, r *http.Request) (time.Time, time.Time, string, *time.Location, bool) {
	zone := strings.TrimSpace(r.URL.Query().Get("timeZone"))
	if zone == "" {
		zone = defaultActivityTimezone
	}
	location, err := time.LoadLocation(zone)
	if err != nil {
		s.writeError(w, r, http.StatusUnprocessableEntity, "INVALID_TIMEZONE", "Use a valid IANA timezone.")
		return time.Time{}, time.Time{}, "", nil, false
	}
	today := time.Now().In(location)
	toDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, location)
	fromDate := toDate.AddDate(0, 0, -29)
	if value := strings.TrimSpace(r.URL.Query().Get("from")); value != "" {
		fromDate, err = time.ParseInLocation(time.DateOnly, value, location)
		if err != nil {
			s.writeError(w, r, http.StatusUnprocessableEntity, "INVALID_DATE_RANGE", "Dates must use YYYY-MM-DD.")
			return time.Time{}, time.Time{}, "", nil, false
		}
	}
	if value := strings.TrimSpace(r.URL.Query().Get("to")); value != "" {
		toDate, err = time.ParseInLocation(time.DateOnly, value, location)
		if err != nil {
			s.writeError(w, r, http.StatusUnprocessableEntity, "INVALID_DATE_RANGE", "Dates must use YYYY-MM-DD.")
			return time.Time{}, time.Time{}, "", nil, false
		}
	}
	if toDate.Before(fromDate) || toDate.Sub(fromDate) > 731*24*time.Hour {
		s.writeError(w, r, http.StatusUnprocessableEntity, "INVALID_DATE_RANGE", "Choose a range of at most two years.")
		return time.Time{}, time.Time{}, "", nil, false
	}
	bucket := r.URL.Query().Get("bucket")
	if bucket == "" {
		bucket = "daily"
	}
	if bucket != "daily" && bucket != "weekly" {
		s.writeError(w, r, http.StatusUnprocessableEntity, "INVALID_BUCKET", "Bucket must be daily or weekly.")
		return time.Time{}, time.Time{}, "", nil, false
	}
	return fromDate.UTC(), toDate.AddDate(0, 0, 1).UTC(), bucket, location, true
}
