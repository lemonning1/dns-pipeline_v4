package repository

import (
	"testing"
	"time"

	"shared/model"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestEnsureTable(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS dns_queries_v4_local").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS dns_queries_v4").
		WillReturnResult(sqlmock.NewResult(0, 0))
	if err := NewDNSRepo(db).EnsureTable(); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestInsertDNSQuery(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	q := &model.DNSQuery{Domain: "a.com", ClientIP: "1.2.3.4", QType: 1, CreatedAt: time.Now()}
	mock.ExpectExec("INSERT INTO dns_queries_v4").
		WithArgs(sqlmock.AnyArg(), q.Domain, q.ClientIP, q.QType, q.QR, q.RCode, q.Cnamechain, q.ResponseIPs, q.TTL, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	if err := NewDNSRepo(db).InsertDNSQuery(q); err != nil {
		t.Fatal(err)
	}
}

func TestInsertDNSQuery_zeroTime(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectExec("INSERT INTO dns_queries_v4").
		WillReturnResult(sqlmock.NewResult(1, 1))
	if err := NewDNSRepo(db).InsertDNSQuery(&model.DNSQuery{Domain: "b.com"}); err != nil {
		t.Fatal(err)
	}
}

func TestInsertBatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	now := time.Now()
	qs := []*model.DNSQuery{
		{ID: 1, Domain: "a.com", ClientIP: "1.1.1.1", QType: 1, CreatedAt: now},
		{ID: 2, Domain: "b.com", ClientIP: "2.2.2.2", QType: 1, CreatedAt: now},
	}

	mock.ExpectExec("INSERT INTO dns_queries_v4").
		WithArgs(
			1, "a.com", "1.1.1.1", 1, 0, 0, "", "", 0, now,
			2, "b.com", "2.2.2.2", 1, 0, 0, "", "", 0, now,
		).
		WillReturnResult(sqlmock.NewResult(2, 2))

	if err := NewDNSRepo(db).InsertBatch(qs); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestInsertBatch_empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := NewDNSRepo(db).InsertBatch(nil); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
