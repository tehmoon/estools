package rules

type DateTimeConfig struct {
	Minus string `json:"minus"`
	Plus string `json:"plus"`

	// Accepts: second, minute, hour, day, week
	Round string `json:"round"`
	Date string `json:"date"`

	// Defaults to time.RFC3339
	Layout string `json:"layout"`
}
