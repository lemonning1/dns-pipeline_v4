package consumerserver

import (
	"context"
	"errors"
	"testing"
	"time"

	"shared/model"
)

type fakeReader struct {
	calls int
	seq   []struct {
		q   *model.DNSQuery
		err error
	}
}

func (f *fakeReader) Read(time.Duration) (*model.DNSQuery, error) {
	if f.calls >= len(f.seq) {
		return nil, nil
	}
	item := f.seq[f.calls]
	f.calls++
	return item.q, item.err
}

type fakeInserter struct {
	err error
	n   int
}

func (f *fakeInserter) Insert(*model.DNSQuery) error {
	f.n++
	return f.err
}
func (f *fakeInserter) InsertBatch([]*model.DNSQuery) error {
	return f.err
}
func TestRun_stop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := Run(ctx, &fakeReader{}, &fakeInserter{}); err != nil {
		t.Fatal(err)
	}
}

func TestRun_readInsert(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	r := &fakeReader{seq: []struct {
		q   *model.DNSQuery
		err error
	}{
		{err: errors.New("kafka")},
		{q: nil},
		{q: &model.DNSQuery{Domain: "a.com"}},
		{q: &model.DNSQuery{Domain: "b.com"}},
	}}
	ins := &fakeInserter{err: errors.New("db")}
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	if err := Run(ctx, r, ins); err != nil {
		t.Fatal(err)
	}
	if ins.n < 1 {
		t.Fatalf("insert n=%d", ins.n)
	}
}

func TestRun_insertOK(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	r := &fakeReader{seq: []struct {
		q   *model.DNSQuery
		err error
	}{
		{q: &model.DNSQuery{Domain: "ok.com"}},
	}}
	ins := &fakeInserter{}
	go func() {
		time.Sleep(40 * time.Millisecond)
		cancel()
	}()
	if err := Run(ctx, r, ins); err != nil {
		t.Fatal(err)
	}
	if ins.n < 1 {
		t.Fatalf("n=%d", ins.n)
	}
}
