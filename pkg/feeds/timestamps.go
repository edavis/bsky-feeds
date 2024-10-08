package feeds

import (
	"time"
)

func SafeTimestamp(input string) int64 {
	utcNow := time.Now().UTC()
	if input == "" {
		return utcNow.Unix()
	}
	var t time.Time
	var err error
	layouts := []string{
		time.RFC3339,
	}
	for _, layout := range layouts {
		if t, err = time.Parse(layout, input); err == nil {
			break
		}
	}
	if err != nil {
		return utcNow.Unix()
	}
	if t.Unix() <= 0 {
		return utcNow.Unix()
	} else if t.Add(-2*time.Minute).Compare(utcNow) == -1 {
		// accept as long as parsed time is no more than 2 minutes in the future
		return t.Unix()
	} else if t.Compare(utcNow) == 1 {
		return utcNow.Unix()
	} else {
		return utcNow.Unix()
	}
}
