package grpcx

import (
	"net"
	"time"
)

const (
	CfgDefaultHost                  = "localhost"
	CfgDefaultHostKube              = "0.0.0.0"
	CfgDefaultPort                  = "50051"
	CfgDefaultReflection            = false
	CfgDefaultMaxConnectionIdle     = 5 * time.Minute
	CfgDefaultMaxConnectionAge      = 2 * time.Hour
	CfgDefaultMaxConnectionAgeGrace = 5 * time.Minute
	CfgDefaultKeepAliveTime         = 2 * time.Hour
	CfgDefaultKeepAliveTimeout      = 20 * time.Second
	CfgDefaultMaxRecvMsgSize        = 1024 * 1024 * 20
	CfgDefaultMaxSendMsgSize        = 1024 * 1024 * 20
	CfgDefaultTLSCertFile           = ""
	CfgDefaultTLSKeyFile            = ""
)

type Config struct {
	Host                  string
	Port                  string
	Reflection            bool
	MaxConnectionIdle     time.Duration
	MaxConnectionAge      time.Duration
	MaxConnectionAgeGrace time.Duration
	KeepAliveTime         time.Duration
	KeepAliveTimeout      time.Duration
	MaxRecvMsgSize        int
	MaxSendMsgSize        int
	TLSCertFile           string
	TLSKeyFile            string
}

// Addr returns server address in format ":<port>".
func (c Config) Addr() string {
	return net.JoinHostPort(c.Host, c.Port)
}

func (c Config) AddrServer() string {
	return net.JoinHostPort(CfgDefaultHostKube, c.Port)
}

func CfgDefault(port string) *Config {
	return &Config{
		Host:                  CfgDefaultHost,
		Port:                  port,
		Reflection:            CfgDefaultReflection,
		MaxConnectionIdle:     CfgDefaultMaxConnectionIdle,
		MaxConnectionAge:      CfgDefaultMaxConnectionAge,
		MaxConnectionAgeGrace: CfgDefaultMaxConnectionAgeGrace,
		KeepAliveTime:         CfgDefaultKeepAliveTime,
		KeepAliveTimeout:      CfgDefaultKeepAliveTimeout,
		MaxRecvMsgSize:        CfgDefaultMaxRecvMsgSize,
		MaxSendMsgSize:        CfgDefaultMaxSendMsgSize,
		TLSCertFile:           CfgDefaultTLSCertFile,
		TLSKeyFile:            CfgDefaultTLSKeyFile,
	}
}
