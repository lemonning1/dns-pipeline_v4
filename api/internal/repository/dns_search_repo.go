package repository

import (
	"database/sql"
	"shared/model"
)

type DNSRepo struct {
	db *sql.DB
}

func NewDNSRepo(db *sql.DB) *DNSRepo {
	return &DNSRepo{db: db}
}

func (d *DNSRepo) FindByDomain(domain string, qr *int, page model.PageParams) ([]model.DNSQuery, int, error) {
	where := " WHERE 1=1 "
	args := []any{}
	if domain != "" {
		where += " AND domain LIKE ? "
		args = append(args, "%"+domain+"%")
	}
	if qr != nil {
		where += " AND qr = ? "
		args = append(args, *qr)
	}

	var total uint64
	countSQL := "SELECT count() FROM dns_queries_v4" + where
	if err := d.db.QueryRow(countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := page.GetOffset()
	limit := page.PageSize

	listSQL := `
		SELECT id, domain, qtype, qr, rcode, cnamechain, responseips, ttl, created_at
		FROM dns_queries_v4` + where + `
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`

	listArgs := append(append([]any{}, args...), limit, offset)

	rows, err := d.db.Query(listSQL, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []model.DNSQuery
	for rows.Next() {
		var q model.DNSQuery
		err := rows.Scan(&q.ID, &q.Domain, &q.QType, &q.QR,
			&q.RCode, &q.Cnamechain, &q.ResponseIPs, &q.TTL, &q.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		results = append(results, q)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	if results == nil {
		results = []model.DNSQuery{}
	}
	return results, int(total), nil
}

func (d *DNSRepo) TopDomains(ddl string) ([]model.TopResult, error) {

	topdomain := `SELECT domain,
		count() AS cnt 
		FROM dns_queries_v4 
		WHERE created_at >= now() -INTERVAL ` + ddl + ` HOUR 
		AND qr = 0 
		GROUP BY domain 
		ORDER BY cnt DESC
		LIMIT 20;`
	rows, err := d.db.Query(topdomain)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []model.TopResult
	for rows.Next() {
		var result model.TopResult
		if err := rows.Scan(&result.Name, &result.Count); err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if results == nil {
		results = []model.TopResult{}
	}
	return results, nil

}
