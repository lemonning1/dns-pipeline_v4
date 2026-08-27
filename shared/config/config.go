package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"go.yaml.in/yaml/v3"
)

type Config struct {
	API      APIConfig      `yaml:"api"`
	Pipeline PipelineConfig `yaml:"pipeline"`
}

type APIConfig struct {
	Database DatabaseConfig  `yaml:"database"`
	API      APIServerConfig `yaml:"api"`
	Log      LogConfig       `yaml:"log"`
}

type PipelineConfig struct {
	Database    DatabaseConfig  `yaml:"database"`
	Kafka       KafkaConfig     `yaml:"kafka"`
	Collector   CollectorConfig `yaml:"collector"`
	ProducerLog LogConfig       `yaml:"producer_log"`
	ConsumerLog LogConfig       `yaml:"consumer_log"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"-"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

type KafkaConfig struct {
	Brokers         []string `yaml:"brokers"`
	Topic           string   `yaml:"topic"`
	GroupID         string   `yaml:"group_id"`
	AutoOffsetReset string   `yaml:"auto_offset_reset"`
}

type CollectorConfig struct {
	Device  string `yaml:"device"`
	SnapLen int32  `yaml:"snaplen"`
	Filter  string `yaml:"filter"`
}

type LogConfig struct {
	Level      string `yaml:"level"`
	File       string `yaml:"file"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
}

type APIServerConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
	Mode string `yaml:"mode"`
}

func (d DatabaseConfig) DSN() string {
	host := d.Host
	if !strings.Contains(host, ",") {
		host = fmt.Sprintf("%s:%d", d.Host, d.Port)
	}

	u := url.URL{
		Scheme: "clickhouse",
		User:   url.UserPassword(d.User, d.Password),
		Host:   host,
		Path:   "/" + d.DBName,
	}
	return u.String()
}

func Load(p string) (*Config, error) {
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	pwd := os.Getenv("DB_PASSWORD")
	if pwd == "" {
		return nil, fmt.Errorf("环境变量 DB_PASSWORD 未设置")
	}
	cfg.API.Database.Password = pwd
	cfg.Pipeline.Database.Password = pwd

	return &cfg, nil
}
