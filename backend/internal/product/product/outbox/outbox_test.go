package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
)

func TestMessagePreservesExactJSON(t *testing.T) {
	amount := "9007199254740993.123456789"
	message, err := NewMessage("when-product-updated", "scope", map[string]string{"amount": amount})
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(message)
	if err != nil {
		t.Fatal(err)
	}
	var restored Message
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	if string(restored.Payload) != string(message.Payload) || !strings.Contains(string(data), amount) {
		t.Fatal("exact value changed")
	}
	if _, err := NewMessage("", "scope", nil); err == nil {
		t.Fatal("blank topic accepted")
	}
	if _, err := NewMessage("topic", "", nil); err == nil {
		t.Fatal("blank key accepted")
	}
	if _, err := NewMessage("topic", "key", make(chan int)); err == nil {
		t.Fatal("unsupported payload accepted")
	}
}

func TestAggregateKeySeparatesCompanyAndIdentity(t *testing.T) {
	a := AggregateKey("holding", "company-a", "product")
	if a == AggregateKey("holding", "company-b", "product") {
		t.Fatal("company collision")
	}
	if AggregateKey("a/b", "c", "d") == AggregateKey("a", "b/c", "d") {
		t.Fatal("separator collision")
	}
	if a != AggregateKey("holding", "company-a", "product") {
		t.Fatal("unstable key")
	}
}

type scriptedPersister struct {
	results []error
	calls   int
}

func (p *scriptedPersister) Transaction(_ context.Context, fn func(context.Context) error) error {
	p.calls++
	if len(p.results) == 0 {
		return fn(context.Background())
	}
	err := p.results[0]
	p.results = p.results[1:]
	if err == nil {
		return fn(context.Background())
	}
	return err
}

func (p *scriptedPersister) Exec(context.Context, interface{}) (*mongo.Collection, error) {
	return nil, errors.New("not used")
}

// Two writers on the same aggregate make MongoDB abort one transaction with the
// TransientTransactionError label; the intent must be replayed, not surfaced as
// an internal error. Other failures and exhausted retries surface unchanged.
func TestRunTransactionRetriesOnlyTransientConflicts(t *testing.T) {
	transient := mongo.CommandError{Code: 112, Message: "WriteConflict", Labels: []string{"TransientTransactionError"}}
	duplicate := mongo.CommandError{Code: 11000, Message: "E11000 duplicate key"}

	pst := &scriptedPersister{results: []error{transient, transient, nil}}
	ran := 0
	if err := runTransaction(context.Background(), pst, func(context.Context) error { ran++; return nil }); err != nil || pst.calls != 3 || ran != 1 {
		t.Fatalf("transient conflict not replayed: err=%v calls=%d ran=%d", err, pst.calls, ran)
	}

	pst = &scriptedPersister{results: []error{duplicate}}
	if err := runTransaction(context.Background(), pst, func(context.Context) error { return nil }); err == nil || isTransientTransactionError(err) || pst.calls != 1 {
		t.Fatalf("non-transient error must not be retried: err=%v calls=%d", err, pst.calls)
	}

	pst = &scriptedPersister{results: []error{transient, transient, transient, transient, transient, transient}}
	if err := runTransaction(context.Background(), pst, func(context.Context) error { return nil }); !isTransientTransactionError(err) || pst.calls != transientTransactionAttempts {
		t.Fatalf("exhausted retries must surface the conflict: err=%v calls=%d", err, pst.calls)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	pst = &scriptedPersister{results: []error{transient, nil}}
	if err := runTransaction(ctx, pst, func(context.Context) error { return nil }); !isTransientTransactionError(err) || pst.calls != 1 {
		t.Fatalf("cancelled request must not keep retrying: err=%v calls=%d", err, pst.calls)
	}
}
