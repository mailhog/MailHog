package api

import (
	"net/http"
	"testing"
	"time"
)

func TestParseDeleteOlderThanParsesHourDuration(t *testing.T) {
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	cutoff, err := parseDeleteOlderThan("1h", now)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := cutoff, now.Add(-time.Hour); !got.Equal(want) {
		t.Fatalf("got cutoff %s, want %s", got, want)
	}
}

func TestParseDeleteOlderThanParsesDayAndWeekDurations(t *testing.T) {
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	cutoff, err := parseDeleteOlderThan("1d", now)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := cutoff, now.Add(-24*time.Hour); !got.Equal(want) {
		t.Fatalf("got cutoff %s, want %s", got, want)
	}

	cutoff, err = parseDeleteOlderThan("week", now)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := cutoff, now.Add(-7*24*time.Hour); !got.Equal(want) {
		t.Fatalf("got cutoff %s, want %s", got, want)
	}
}

func TestParseDeleteOlderThanRejectsTimestamps(t *testing.T) {
	now := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	if _, err := parseDeleteOlderThan("1700000000000", now); err == nil {
		t.Fatal("expected timestamp-like older_than value to be rejected")
	}
}

func TestDeleteOlderThanCutoffParsesCreatedBefore(t *testing.T) {
	req, err := http.NewRequest("DELETE", "/api/v3/messages?created_before=2026-05-22T12:34:56Z", nil)
	if err != nil {
		t.Fatal(err)
	}

	cutoff, ok, err := deleteOlderThanCutoff(req)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected cutoff to be present")
	}
	if got, want := cutoff.Format(time.RFC3339), "2026-05-22T12:34:56Z"; got != want {
		t.Fatalf("got cutoff %s, want %s", got, want)
	}
}

func TestDeleteOlderThanCutoffParsesCreatedBeforeUnixMilliseconds(t *testing.T) {
	req, err := http.NewRequest("DELETE", "/api/v3/messages?created_before=1700000000000", nil)
	if err != nil {
		t.Fatal(err)
	}

	cutoff, ok, err := deleteOlderThanCutoff(req)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected cutoff to be present")
	}
	if got, want := cutoff, time.Unix(0, 1700000000000*int64(time.Millisecond)); !got.Equal(want) {
		t.Fatalf("got cutoff %s, want %s", got, want)
	}
}

func TestDeleteOlderThanCutoffRejectsMixedFilters(t *testing.T) {
	req, err := http.NewRequest("DELETE", "/api/v3/messages?older_than=1h&created_before=2026-05-22T12:34:56Z", nil)
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = deleteOlderThanCutoff(req)
	if err == nil {
		t.Fatal("expected mixed filter error")
	}
}

func TestDeleteOlderThanCutoffRejectsInvalidDuration(t *testing.T) {
	req, err := http.NewRequest("DELETE", "/api/v3/messages?older_than=not-a-date", nil)
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = deleteOlderThanCutoff(req)
	if err == nil {
		t.Fatal("expected invalid cutoff error")
	}
}
