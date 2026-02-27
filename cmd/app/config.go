package main

// Config ...
type Config struct {
	DatabaseURL     string `toml:"database_url"`
	TestDatabaseURL string `toml:"test_database_url"`
	RedisAddr       string `toml:"redis_addr"`
	TestRedisAddr   string `toml:"test_redis_addr"`
	BindAddr        string `toml:"bind_addr"`
	LogLevel        string `toml:"log_level"`
	JWTSecret       string `toml:"jwt_secret"`
	AuthServiceAddr string `toml:"auth_service_addr"`
}

// NewConfig ...
func NewConfig() *Config {
	return &Config{
		BindAddr: ":8080",
		LogLevel: "info",
	}
}
