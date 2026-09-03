package consumerserver

import (
	"context"
	"errors"
	"testing"
	"time"

	"pipeline/internal/kafka"
	"shared/model"

	confluent "github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type fakeKafkaClient struct {
	calls    int
	seq      []readResult
	commits  int
	commitErr error
}

type readResult struct {
	cm  *kafka.ConsumedMessage
	err error
}

func (f *fakeKafkaClient) Read(time.Duration) (*kafka.ConsumedMessage, error) {
	if f.calls >= len(f.seq) {
		return nil, nil
	}
	item := f.seq[f.calls]
	f.calls++
	return item.cm, item.err
}

func (f *fakeKafkaClient) Commit(*confluent.Message) error {
	f.commits++
	return f.commitErr
}

func fakeCM(domain string, partition int32, offset int64) *kafka.ConsumedMessage {
	topic := "dns_topic"
	return &kafka.ConsumedMessage{
		Record: &model.DNSQuery{Domain: domain},
		KafkaMsg: &confluent.Message{
			TopicPartition: confluent.TopicPartition{
				Topic:     &topic,
				Partition: partition,
				Offset:    confluent.Offset(offset),
			},
		},
	}
}

type fakeInserter struct {
	err        error
	batchCalls int
	batchTotal int
}

func (f *fakeInserter) Insert(*model.DNSQuery) error {
	return f.err
}

func (f *fakeInserter) InsertBatch(queries []*model.DNSQuery) error {
	f.batchCalls++
	f.batchTotal += len(queries)
	return f.err
}

func TestRun_stop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := Run(ctx, &fakeKafkaClient{}, &fakeInserter{}); err != nil {
		t.Fatal(err)
	}
}

func TestRun_readInsert(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	r := &fakeKafkaClient{seq: []readResult{
		{err: errors.New("kafka")},
		{cm: nil},
		{cm: fakeCM("a.com", 0, 1)},
		{cm: fakeCM("b.com", 0, 2)},
	}}
	ins := &fakeInserter{err: errors.New("db")}
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	if err := Run(ctx, r, ins); err != nil {
		t.Fatal(err)
	}
	if ins.batchCalls < 1 {
		t.Fatalf("batchCalls=%d", ins.batchCalls)
	}
	if ins.batchTotal != 2 {
		t.Fatalf("batchTotal=%d want 2", ins.batchTotal)
	}
}

func TestRun_insertOKAndCommit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	client := &fakeKafkaClient{seq: []readResult{
		{cm: fakeCM("ok.com", 1, 10)},
	}}
	ins := &fakeInserter{}
	go func() {
		time.Sleep(40 * time.Millisecond)
		cancel()
	}()
	if err := Run(ctx, client, ins); err != nil {
		t.Fatal(err)
	}
	if ins.batchTotal != 1 {
		t.Fatalf("batchTotal=%d want 1", ins.batchTotal)
	}
	if client.commits != 1 {
		t.Fatalf("commits=%d want 1", client.commits)
	}
}
