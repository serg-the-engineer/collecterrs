package logs

const (
	LevelInfo       = "info"    // Уровень логирования 'info'.
	CfgDefaultLevel = LevelInfo // Уровень логирования по умолчанию.
)

// Config параметры конфигурации логера.
type Config struct {
	Level  string // Уровень логирования.
	Pretty bool   // Форматированный вывод логов.
}

// CfgDefault загружает конфиг с помощью viper.
func CfgDefault() *Config {
	return &Config{
		Level:  CfgDefaultLevel,
		Pretty: true,
	}
}
