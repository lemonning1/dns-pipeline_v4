package repository

import (
	"database/sql"
	"shared/model"
	"strings"
	"time"
)

type DNSRepo struct {
	db *sql.DB
}

func NewDNSRepo(db *sql.DB) *DNSRepo {
	return &DNSRepo{db: db}
}

func (d *DNSRepo) EnsureTable() error {
	_, err := d.db.Exec(`
		CREATE TABLE IF NOT EXISTS dns_queries_v4_local ON CLUSTER dns_cluster
		(
			id UInt64,
			domain String,
			clientip String,
			qtype Int32,
			qr Int32,
			rcode Int32,
			cnamechain String,
			responseips String,
			ttl Int32,
			created_at DateTime
		)
		ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/dns_queries_v4', '{replica}')
		ORDER BY (created_at, domain)
	`)
	if err != nil {
		return err
	}

	_, err = d.db.Exec(`
		CREATE TABLE IF NOT EXISTS dns_queries_v4 ON CLUSTER dns_cluster
		AS dns_queries_v4_local
		ENGINE = Distributed('dns_cluster', currentDatabase(), 'dns_queries_v4_local', rand())
	`)
	return err
}

func (d *DNSRepo) InsertDNSQuery(query *model.DNSQuery) error {
	createdAt := query.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	id := query.ID
	if id == 0 {
		id = int(time.Now().UnixNano() & 0x7fffffffffffffff)
	}
	_, err := d.db.Exec(`
		INSERT INTO dns_queries_v4 (id, domain, clientip, qtype, qr, rcode, cnamechain, responseips, ttl, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id, query.Domain, query.ClientIP, query.QType, query.QR, query.RCode, query.Cnamechain, query.ResponseIPs, query.TTL, createdAt)
	return err
}
func (d *DNSRepo) InsertBatch(queries []*model.DNSQuery) error {
	if len(queries) == 0 {
		return nil
	}

	placeholders := make([]string, 0, len(queries))
	args := make([]any, 0, len(queries)*10)

	for _, query := range queries {
		createdAt := query.CreatedAt
		if createdAt.IsZero() {
			createdAt = time.Now()
		}
		id := query.ID
		if id == 0 {
			// Consumer 侧已用 Kafka offset 赋 id；仅单测或未走 kafka 时兜底。
			id = int(time.Now().UnixNano() & 0x7fffffffffffffff)
		}

		placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
		args = append(args,
			id, query.Domain, query.ClientIP, query.QType, query.QR, query.RCode,
			query.Cnamechain, query.ResponseIPs, query.TTL, createdAt,
		)
	}

	sqlStr := `
		INSERT INTO dns_queries_v4
		(id, domain, clientip, qtype, qr, rcode, cnamechain, responseips, ttl, created_at)
		VALUES ` + strings.Join(placeholders, ",")

	_, err := d.db.Exec(sqlStr, args...)
	return err
}
