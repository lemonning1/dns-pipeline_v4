package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDatabaseConfigDSN(t *testing.T) {
	d := DatabaseConfig{
		Host:     "127.0.0.1",
		Port:     9000,
		User:     "default",
		Password: "secret",
		DBName:   "dnsdb",
	}
	got := d.DSN()
	want := "clickhouse://default:secret@127.0.0.1:9000/dnsdb"
	if got != want {
		t.Fatalf("DSN()=\n%s\nwant\n%s", got, want)
	}
}

func TestLoad_fileMissing(t *testing.T) {
	t.Setenv("DB_PASSWORD", "x")
	_, err := Load(filepath.Join(t.TempDir(), "nope.yaml"))
	if err == nil || !strings.Contains(err.Error(), "读取配置文件失败") {
		t.Fatalf("err=%v", err)
	}
}

func TestLoad_badYAML(t *testing.T) {
	t.Setenv("DB_PASSWORD", "x")
	p := filepath.Join(t.TempDir(), "bad.yaml")
	if err := os.WriteFile(p, []byte(":\n  - ["), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(p)
	if err == nil || !strings.Contains(err.Error(), "解析配置文件失败") {
		t.Fatalf("err=%v", err)
	}
}

func TestLoad_missingPassword(t *testing.T) {
	t.Setenv("DB_PASSWORD", "")
	p := filepath.Join(t.TempDir(), "ok.yaml")
	if err := os.WriteFile(p, []byte("pipeline:\n  kafka:\n    topic: t\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(p)
	if err == nil || !strings.Contains(err.Error(), "DB_PASSWORD") {
		t.Fatalf("err=%v", err)
	}
}

func TestLoad_ok(t *testing.T) {
	t.Setenv("DB_PASSWORD", "pw")
	p := filepath.Join(t.TempDir(), "ok.yaml")
	body := `
api:
  database:
    host: h
    port: 1
    user: u
    dbname: d
  api:
    host: 0.0.0.0
    port: "8080"
pipeline:
  kafka:
    topic: dns_topic
    group_id: g
  producer_log:
    level: debug
    file: logs/p.log
`
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.API.Database.Password != "pw" || cfg.Pipeline.Database.Password != "pw" {
		t.Fatalf("password not applied: %+v", cfg.API.Database)
	}
	if cfg.Pipeline.Kafka.Topic != "dns_topic" {
		t.Fatalf("topic=%s", cfg.Pipeline.Kafka.Topic)
	}
	if cfg.Pipeline.ProducerLog.Level != "debug" {
		t.Fatalf("log=%+v", cfg.Pipeline.ProducerLog)
	}
}
