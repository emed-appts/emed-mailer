package collector

// DBConfig struct encapsulate all settings for dbCollector
type DBConfig struct {
	Server   string
	Port     int
	User     string
	Password string
	Database string
}
