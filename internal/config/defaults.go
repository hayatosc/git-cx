package config

// DefaultConfig returns a Config populated with default values.
func DefaultConfig() *Config {
	return &Config{
		Provider:   "gemini",
		Model:      "gemini-3.0-flash",
		Candidates: 3,
		Timeout:    30,
		API: APIConfig{
			BaseURL: "https://api.openai.com/v1",
		},
		Commit: CommitConfig{
			UseEmoji:         false,
			MaxSubjectLength: 100,
		},
	}
}
