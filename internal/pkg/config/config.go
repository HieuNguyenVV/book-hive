package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	App       AppConfig      `mapstructure:"app"`
	Postgres  PostgresConfig `mapstructure:"postgres"`
	Redis     RedisConfig    `mapstructure:"redis"`
	JWT       JWTConfig      `mapstructure:"jwt"`
	Log       LogConfig      `mapstructure:"log"`
	WebSocket WebSocket      `mapstructure:"web_socket"`
}

type AppConfig struct {
	Env        string `mapstructure:"env"`
	Port       string `mapstructure:"port"`
	Host       string `mapstructure:"host"`
	Debug      bool   `mapstructure:"debug"`
	ApiVersion string `mapstructure:"api_version"`
	ApiPrefix  string `mapstructure:"api_prefix"`
}

type LogConfig struct {
	LogLevel string `mapstructure:"log_level"`
}

type PostgresConfig struct {
	Master PostgresNodeConfig `mapstructure:"master"`
	Slaves PostgresNodeConfig `mapstructure:"slaves"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type PostgresNodeConfig struct {
	Host         string `mapstructure:"host"`
	Port         string `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	Database     string `mapstructure:"database"`
	SSLMode      string `mapstructure:"ssl_mode"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}
type JWTConfig struct {
	AccessTokenSecret string `mapstructure:"access_token_secret"`
	AccessTokenTTL    int    `mapstructure:"access_token_ttl"`
	RefreshTokenTTL   int    `mapstructure:"refresh_token_ttl"`
}

type WebSocket struct {
	// Time to wait for a write response (default: 10 seconds)
	WriteWaitSec int `mapstructure:"write_wait_sec"`
	// Time to wait for a pong response (default: 60 seconds)
	PongWaitSec int `mapstructure:"pong_wait_sec"`
	// Time to wait for a ping response (default: 55 seconds)
	PingPeriodSec int `mapstructure:"ping_period_sec"`
	// Maximum number of connections (default: 1000)
	MaxConnections int `mapstructure:"max_connections"`
	// Enable concurrent message handling (default: false)
	ConcurrentMessageHandling bool `mapstructure:"concurrent_message_handling"`
	// Maximum number of workers for dispatching messages (default: 10)
	MaxDispatchWorkers int `mapstructure:"max_dispatch_workers"`
	// Timeout for notifying clients about new messages (default: 1000ms)
	ClientNotifyTimeoutSec int `mapstructure:"client_notify_timeout_sec"`
	// Buffer size for the message channel (default: 1000)
	MessageChannelBufferSize int `mapstructure:"message_channel_buffer_size"`
	// Timeout for sending messages to the message channel (default: 1000ms)
	MessageChannelSendTimeoutMs int `mapstructure:"message_channel_send_timeout_ms"`
	// Enable message rate limiting (default: false)
	MesgRateLimitEnabled bool `mapstructure:"mesg_rate_limit_enabled"`
	// Messages per second allowed for a single client (default: 100)
	MesgRateLimitPerSecond int `mapstructure:"mesg_rate_limit_per_second"`
	// Max retry attempts for rate limiting before dropping the connection
	MesgRateLimitMaxRetry int `mapstructure:"mesg_rate_limit_max_retry"`
	// Max message size in bytes (default: 1024 * 1024 * 10 = 10MB)
	MaxMessageSize int64 `mapstructure:"max_message_size"`
}

func LoadConfig() (*Config, error) {
	return loadConfig([]string{"."})
}

func loadConfig(paths []string) (*Config, error) {
	var config Config

	_ = godotenv.Load()
	for _, path := range paths {
		f := filepath.Join(path, ".env")
		if _, err := os.Stat(f); err == nil {
			_ = godotenv.Load(f)
		}
	}

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	for _, path := range paths {
		v.AddConfigPath(path)
	}

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return &config, nil
}

func (c PostgresNodeConfig) DSN() string {
	sslMode := c.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Password),
		Host:   fmt.Sprintf("%s:%s", c.Host, c.Port),
		Path:   c.Database,
	}
	q := u.Query()
	q.Set("sslmode", sslMode)
	u.RawQuery = q.Encode()

	return u.String()
}
