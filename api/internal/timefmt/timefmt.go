package timefmt

import "time"

const Layout = "2006-01-02 15:04:05"

var parseLayouts = []string{Layout, time.RFC3339Nano, time.RFC3339}

func NowUTC() string {
	return time.Now().UTC().Format(Layout)
}

func FormatUTC(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(Layout)
}

func Parse(s string) (time.Time, bool) {
	for _, layout := range parseLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func Normalize(s string) string {
	if s == "" {
		return ""
	}
	if t, ok := Parse(s); ok {
		return FormatUTC(t)
	}
	return s
}
