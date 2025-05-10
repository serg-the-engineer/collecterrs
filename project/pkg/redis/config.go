package redis

const (
	CfgDefaultHost               = "127.0.0.1"
	CfgDefaultPort               = "6379"
	CfgDefaultPassword           = ""
	CfgDefaultCertCa             = ""
	CfgDefaultDatabase           = 1
	CfgDefaultMasterName         = ""
	CfgDefaultClusterEnabled     = false
	CfgDefaultMaxBulkRequestSize = 10
)

type (
	Config struct {
		Host               string
		Port               string
		Database           int
		Password           string
		CertCA             string // Локальный путь до сертификата.
		MasterName         string
		ClusterEnabled     bool
		MaxBulkRequestSize int
	}
)

func NewConfig(
	host, port, password, certCa, masterName string,
	database int,
	clusterEnabled bool,
	bulkSize int,
) *Config {
	return &Config{
		Host:               host,
		Port:               port,
		Database:           database,
		Password:           password,
		CertCA:             certCa,
		MasterName:         masterName,
		ClusterEnabled:     clusterEnabled,
		MaxBulkRequestSize: bulkSize,
	}
}

func CfgDefault() *Config {
	return NewConfig(
		CfgDefaultHost,
		CfgDefaultPort,
		CfgDefaultPassword,
		CfgDefaultCertCa,
		CfgDefaultMasterName,
		CfgDefaultDatabase,
		CfgDefaultClusterEnabled,
		CfgDefaultMaxBulkRequestSize,
	)
}
