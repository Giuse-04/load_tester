package config

type AttackConfig struct {
	URL           string
	Concurrency   int
	TotalRequests int
}

func DefaultConfig() AttackConfig {
	return AttackConfig{
		URL:           "http://google.com",
		Concurrency:   10,
		TotalRequests: 100,
	}
}

type Request struct {
	ID      int
	Method  string
	URL     string
	Headers map[string]string
	Body    []byte
}
