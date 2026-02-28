package config

// DefaultConfig returns a Config populated with default values.
func DefaultConfig() *Config {
	return &Config{
		Provider:   "gemini",
		Model:      "gemini-2.0-flash",
		Candidates: 3,
		Timeout:    30,
		Commit: CommitConfig{
			UseEmoji:         false,
			MaxSubjectLength: 100,
		},
	}
}
