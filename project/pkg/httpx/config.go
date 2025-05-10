package httpx

import (
	"net"
	"time"
)

const (
	CfgDefaultPort              = "8080"
	CfgDefaultReadHeaderTimeout = 30 * time.Second
	CfgDefaultTimeout           = 3000 * time.Millisecond
	CfgDefaultCacheTTL          = 300 * time.Second
	CfgDefaultBaseURL           = "/api/"
	CfgDefaultMaxRate           = 10
	CfgDefaultRatePeriod        = 1 * time.Minute
	CfgDefaultCORSOrigins       = "*"
	CfgDefaultCORSMethod        = "GET,POST,PUT,DELETE,OPTIONS"
	CfgDefaultCORSHeaders       = "*"
)

type Config struct {
	Port              string
	ReadHeaderTimeout time.Duration
	Timeout           time.Duration
	CacheTTL          time.Duration
	BaseURL           string
	Rate              RateLimit
	CORS              CORSConfig
}

type RateLimit struct {
	MaxRate int
	Every   time.Duration
}

type CORSConfig struct {
	Origins     string
	Methods     string
	Headers     string
	Credentials bool
}

// Addr returns server address in format ":<port>".
func (c Config) Addr() string {
	return net.JoinHostPort("", c.Port)
}

func CfgDefault() *Config {
	return &Config{
		Port:              CfgDefaultPort,
		ReadHeaderTimeout: CfgDefaultReadHeaderTimeout,
		Timeout:           CfgDefaultTimeout,
		CacheTTL:          CfgDefaultCacheTTL,
		BaseURL:           CfgDefaultBaseURL,
		Rate: RateLimit{
			MaxRate: CfgDefaultMaxRate,
			Every:   CfgDefaultRatePeriod,
		},
		CORS: CORSConfig{
			Origins:     CfgDefaultCORSOrigins,
			Methods:     CfgDefaultCORSMethod,
			Headers:     CfgDefaultCORSHeaders,
			Credentials: false,
		},
	}
}
