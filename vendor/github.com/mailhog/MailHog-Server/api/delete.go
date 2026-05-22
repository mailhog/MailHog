package api

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mailhog/storage"
)

func deleteMessages(storageBackend storage.Storage, req *http.Request) (int, error) {
	cutoff, ok, err := deleteOlderThanCutoff(req)
	if err != nil {
		return http.StatusBadRequest, err
	}
	if ok {
		return http.StatusOK, storageBackend.DeleteOlderThan(cutoff)
	}
	return http.StatusOK, storageBackend.DeleteAll()
}

func deleteOlderThanCutoff(req *http.Request) (time.Time, bool, error) {
	query := req.URL.Query()
	olderThan := strings.TrimSpace(query.Get("older_than"))
	createdBefore := strings.TrimSpace(query.Get("created_before"))

	if olderThan != "" && createdBefore != "" {
		return time.Time{}, false, fmt.Errorf("older_than and created_before cannot be used together")
	}

	if olderThan != "" {
		cutoff, err := parseDeleteOlderThan(olderThan, time.Now())
		return cutoff, true, err
	}

	if createdBefore != "" {
		cutoff, err := parseCreatedBefore(createdBefore)
		return cutoff, true, err
	}

	return time.Time{}, false, nil
}

var olderThanPattern = regexp.MustCompile(`^([0-9]+)\s*([a-z]+)$`)

func parseDeleteOlderThan(value string, now time.Time) (time.Time, error) {
	if value == "" {
		return time.Time{}, fmt.Errorf("older_than value is empty")
	}

	duration, err := parseOlderThanDuration(value)
	if err != nil {
		return time.Time{}, err
	}
	return now.Add(-duration), nil
}

func parseOlderThanDuration(value string) (time.Duration, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return 0, fmt.Errorf("older_than value is empty")
	}

	switch normalized {
	case "hour", "hours":
		return time.Hour, nil
	case "day", "days":
		return 24 * time.Hour, nil
	case "week", "weeks":
		return 7 * 24 * time.Hour, nil
	}

	if duration, err := time.ParseDuration(normalized); err == nil {
		if duration <= 0 {
			return 0, fmt.Errorf("older_than must be positive")
		}
		return duration, nil
	}

	match := olderThanPattern.FindStringSubmatch(normalized)
	if match == nil {
		return 0, fmt.Errorf("older_than must be a duration like 1h, 24h, 1d, 7d, or 1w")
	}

	count, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil || count <= 0 {
		return 0, fmt.Errorf("older_than must be positive")
	}

	switch match[2] {
	case "h", "hr", "hrs", "hour", "hours":
		return time.Duration(count) * time.Hour, nil
	case "d", "day", "days":
		return time.Duration(count) * 24 * time.Hour, nil
	case "w", "week", "weeks":
		return time.Duration(count) * 7 * 24 * time.Hour, nil
	}

	return 0, fmt.Errorf("older_than must use h, d, or w units")
}

func parseCreatedBefore(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, fmt.Errorf("created_before value is empty")
	}

	if n, err := strconv.ParseInt(value, 10, 64); err == nil {
		if n <= 0 {
			return time.Time{}, fmt.Errorf("created_before must be positive")
		}
		if len(value) > 10 {
			return time.Unix(0, n*int64(time.Millisecond)), nil
		}
		return time.Unix(n, 0), nil
	}

	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		if cutoff, err := time.Parse(layout, value); err == nil {
			return cutoff, nil
		}
	}

	return time.Time{}, fmt.Errorf("created_before must be Unix seconds, Unix milliseconds, or RFC3339 timestamp")
}
