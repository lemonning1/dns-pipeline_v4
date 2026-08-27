package repository

import (
	"database/sql"
	"shared/model"
	"time"
)

type DNSRepo struct {
	db *sql.DB
}

func NewDNSRepo(db *sql.DB) *DNSRepo {
	return &DNSRepo{db: db}
}

// EnsureTable 在 dns_cluster（1 分片 2 副本）上建本地复制表 + 分布式表。
// 应用读写表名仍是 dns_queries_v4（Distributed）。
func (d *DNSRepo) EnsureTable() error {
	_, err := d.db.Exec(`
		CREATE TABLE IF NOT EXISTS dns_queries_v4_local ON CLUSTER dns_cluster
		(
			id UInt64,
			domain String,
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
		INSERT INTO dns_queries_v4 (id, domain, qtype, qr, rcode, cnamechain, responseips, ttl, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id, query.Domain, query.QType, query.QR, query.RCode, query.Cnamechain, query.ResponseIPs, query.TTL, createdAt)
	return err
}
