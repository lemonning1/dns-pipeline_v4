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

	q := &model.DNSQuery{Domain: "a.com", QType: 1, CreatedAt: time.Now()}
	mock.ExpectExec("INSERT INTO dns_queries_v4").
		WithArgs(sqlmock.AnyArg(), q.Domain, q.QType, q.QR, q.RCode, q.Cnamechain, q.ResponseIPs, q.TTL, sqlmock.AnyArg()).
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
