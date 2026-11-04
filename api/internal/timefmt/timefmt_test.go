package timefmt_test

import (
	"testing"
	"time"

	"github.com/ArminDashti/agent-brain/agent-brain-api/internal/timefmt"
)

func TestNormalizeRFC3339(t *testing.T) {
	got := timefmt.Normalize("2026-04-01T12:30:45.123456789Z")
	want := "2026-04-01 12:30:45"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestFormatUTC(t *testing.T) {
	tm := time.Date(2026, 4, 1, 12, 30, 45, 0, time.UTC)
	if timefmt.FormatUTC(tm) != "2026-04-01 12:30:45" {
		t.Fatal("FormatUTC mismatch")
	}
}
