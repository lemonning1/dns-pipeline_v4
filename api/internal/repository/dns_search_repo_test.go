package repository

import (
	"errors"
	"shared/model"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestFindByDomain_empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewDNSRepo(db)
	page := model.PageParams{Page: 1, PageSize: 20}

	mock.ExpectQuery("SELECT count").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("SELECT id").
		WithArgs(20, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "domain", "qtype", "qr", "rcode", "cnamechain", "responseips", "ttl", "created_at",
		}))

	items, total, err := repo.FindByDomain("", nil, page)
	if err != nil {
		t.Fatal(err)
	}
	if total != 0 || len(items) != 0 {
		t.Fatalf("items=%v total=%d", items, total)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestFindByDomain_filterAndRow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewDNSRepo(db)
	page := model.PageParams{Page: 1, PageSize: 20}
	qr := 1
	created := time.Now()

	mock.ExpectQuery("SELECT count").
		WithArgs("%ex%", 1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT id").
		WithArgs("%ex%", 1, 20, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "domain", "qtype", "qr", "rcode", "cnamechain", "responseips", "ttl", "created_at",
		}).AddRow(7, "ex.com", 1, 1, 0, "c", "1.1.1.1", 30, created))

	items, total, err := repo.FindByDomain("ex", &qr, page)
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(items) != 1 || items[0].Domain != "ex.com" {
		t.Fatalf("items=%v total=%d", items, total)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestFindByDomain_countError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT count").
		WillReturnError(errors.New("count fail"))
	_, _, err = NewDNSRepo(db).FindByDomain("", nil, model.PageParams{Page: 1, PageSize: 20})
	if err == nil {
		t.Fatal("期望错误")
	}
}

func TestFindByDomain_queryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT count").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("SELECT id").
		WillReturnError(errors.New("list fail"))
	_, _, err = NewDNSRepo(db).FindByDomain("", nil, model.PageParams{Page: 1, PageSize: 20})
	if err == nil {
		t.Fatal("期望错误")
	}
}
