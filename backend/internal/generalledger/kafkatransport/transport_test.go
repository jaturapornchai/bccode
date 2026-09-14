package kafkatransport

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	gl "smlcloudplatform/internal/generalledger"
	"smlcloudplatform/internal/product/projection"
)

type testMQ struct{ brokers, protocol, ca, cert, key string }

func (m testMQ) URI() string              { return m.brokers }
func (m testMQ) SecurityProtocol() string { return m.protocol }
func (m testMQ) SSLCAFile() string        { return m.ca }
func (m testMQ) SSLCertFile() string      { return m.cert }
func (m testMQ) SSLKeyFile() string       { return m.key }

func referenceMessage(t *testing.T) kafka.Message {
	t.Helper()
	ref := gl.EventReference{SchemaVersion: 1, EventID: strings.Repeat("a", 64), HoldingCode: "demo", BusinessCode: "C01", Sequence: 1, EventHash: strings.Repeat("b", 64)}
	value, err := json.Marshal(ref)
	if err != nil {
		t.Fatal(err)
	}
	return kafka.Message{Key: []byte(gl.ProjectionKey(ref.HoldingCode, ref.BusinessCode)), Value: value}
}

func TestReferenceEnvelopeIsStrict(t *testing.T) {
	valid := referenceMessage(t)
	if _, err := DecodeReference(valid); err != nil {
		t.Fatal(err)
	}
	for name, value := range map[string]string{
		"unknown payload":     strings.TrimSuffix(string(valid.Value), "}") + `,"payload":{}}`,
		"duplicate field":     strings.TrimSuffix(string(valid.Value), "}") + `,"sequence":2}`,
		"trailing JSON":       string(valid.Value) + `{}`,
		"missing field":       strings.Replace(string(valid.Value), `"sequence":1,`, "", 1),
		"fractional sequence": strings.Replace(string(valid.Value), `"sequence":1`, `"sequence":1.5`, 1),
		"unknown schema":      strings.Replace(string(valid.Value), `"schemaversion":1`, `"schemaversion":2`, 1),
		"null reference":      "null",
		"oversized reference": strings.Repeat(" ", maxReferenceBytes+1),
	} {
		t.Run(name, func(t *testing.T) {
			message := valid
			message.Value = []byte(value)
			if _, err := DecodeReference(message); err == nil {
				t.Fatal("accepted an invalid financial reference")
			}
		})
	}
	valid.Key = []byte(gl.ProjectionKey("demo", "C02"))
	if _, err := DecodeReference(valid); err == nil {
		t.Fatal("accepted another company's partition key")
	}
}

type captureWriter struct {
	messages []kafka.Message
	timeout  time.Duration
	err      error
}

func (w *captureWriter) WriteMessages(ctx context.Context, messages ...kafka.Message) error {
	w.messages = messages
	deadline, ok := ctx.Deadline()
	if ok {
		w.timeout = time.Until(deadline)
	}
	return w.err
}
func (*captureWriter) Close() error { return nil }

func TestPublishSendsReferenceAndPreservesBrokerFailure(t *testing.T) {
	ackFailure := errors.New("broker acknowledgement failed")
	writer := &captureWriter{err: ackFailure}
	bus := &Bus{writer: writer}
	event := gl.Event{ID: strings.Repeat("c", 64), HoldingCode: "demo", BusinessCode: "C01", Sequence: 1, Actor: "test-actor", Changes: []gl.Change{{Payload: `{"debit":"0.10"}`}}}
	if err := bus.Publish(context.Background(), event); !errors.Is(err, ackFailure) {
		t.Fatalf("broker failure was lost: %v", err)
	}
	if writer.timeout <= 0 || writer.timeout > operationTimeout || len(writer.messages) != 1 {
		t.Fatal("publication must have one message and a bounded deadline")
	}
	if _, err := DecodeReference(writer.messages[0]); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(writer.messages[0].Value), "debit") || strings.Contains(string(writer.messages[0].Value), "test-actor") {
		t.Fatal("financial payload or actor leaked into the reference envelope")
	}
}

func TestTransportConfigurationFailsClosed(t *testing.T) {
	for name, cfg := range map[string]testMQ{
		"empty brokers": {},
		"missing port":  {brokers: "localhost"},
		"credentials":   {brokers: "user:password@localhost:9092"},
		"invalid port":  {brokers: "localhost:70000"},
		"SASL":          {brokers: "localhost:9092", protocol: "SASL_SSL"},
		"TLS downgrade": {brokers: "localhost:9092", ca: "ca.pem"},
		"missing CA":    {brokers: "localhost:9092", protocol: "SSL", ca: filepath.Join(t.TempDir(), "missing.pem")},
		"missing key":   {brokers: "localhost:9092", protocol: "SSL", cert: "cert.pem"},
	} {
		t.Run(name, func(t *testing.T) {
			if bus, err := New(cfg, "test-gl-v2"); err == nil {
				bus.Close()
				t.Fatal("accepted unsupported or incomplete transport configuration")
			}
		})
	}
	bus, err := New(testMQ{brokers: "localhost:9092", protocol: "PLAINTEXT"}, "test-gl-v2")
	if err != nil {
		t.Fatal(err)
	}
	defer bus.Close()
	writer := bus.writer.(*kafka.Writer)
	_, hash := writer.Balancer.(*kafka.Hash)
	if writer.RequiredAcks != kafka.RequireAll || writer.Async || writer.AllowAutoTopicCreation || !hash || writer.ReadTimeout != operationTimeout || writer.WriteTimeout != operationTimeout || writer.Topic != Topic {
		t.Fatal("writer must synchronously acknowledge a keyed, preprovisioned topic")
	}
}

func TestTransportTLSUsesLocalCAAndClientPair(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	cert := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "GL test CA"}, NotBefore: time.Now().Add(-time.Minute), NotAfter: time.Now().Add(time.Hour), IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature}
	der, err := x509.CreateCertificate(rand.Reader, cert, cert, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	keyBytes, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	certPath, keyPath := filepath.Join(dir, "ca.pem"), filepath.Join(dir, "client.key")
	if err := os.WriteFile(certPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyPath, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyBytes}), 0600); err != nil {
		t.Fatal(err)
	}
	bus, err := New(testMQ{brokers: "localhost:9092", protocol: "SSL", ca: certPath, cert: certPath, key: keyPath}, "test-gl-v2")
	if err != nil {
		t.Fatal(err)
	}
	defer bus.Close()
	if bus.transport.TLS.MinVersion != tls.VersionTLS12 || bus.transport.TLS.InsecureSkipVerify || len(bus.transport.TLS.Certificates) != 1 {
		t.Fatal("TLS must validate the server and retain the configured client pair")
	}
}

type fakeReader struct {
	messages  []kafka.Message
	committed int
	commitErr error
	closed    bool
	cancel    context.CancelFunc
}

func (r *fakeReader) FetchMessage(context.Context) (kafka.Message, error) {
	if len(r.messages) == 0 {
		return kafka.Message{}, io.EOF
	}
	m := r.messages[0]
	r.messages = r.messages[1:]
	return m, nil
}
func (r *fakeReader) CommitMessages(context.Context, ...kafka.Message) error {
	if r.commitErr == nil {
		r.committed++
	}
	return r.commitErr
}
func (r *fakeReader) Close() error { r.closed = true; r.cancel(); return nil }

func TestConsumerCommitsOnlyAppliedReferences(t *testing.T) {
	for _, mode := range []string{"success", "apply failure", "rejected sentinel", "panic", "invalid envelope", "commit failure"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			first := referenceMessage(t)
			if mode == "invalid envelope" {
				first.Key = []byte("wrong")
			}
			reader := &fakeReader{messages: []kafka.Message{first, referenceMessage(t)}, cancel: cancel}
			if mode == "commit failure" {
				reader.commitErr = errors.New("rebalance")
			}
			bus := &Bus{newReader: func() messageReader { return reader }}
			calls := 0
			bus.Run(ctx, func(context.Context, gl.EventReference) error {
				calls++
				switch mode {
				case "apply failure":
					return errors.New("PostgreSQL unavailable")
				case "rejected sentinel":
					return projection.ErrRejected
				case "panic":
					panic("handler failure")
				}
				return nil
			}, nil)
			if !reader.closed {
				t.Fatal("reader was not closed before shutdown")
			}
			if mode == "success" {
				if reader.committed != 2 || calls != 2 {
					t.Fatal("successful work was not acknowledged")
				}
			} else if reader.committed != 0 || len(reader.messages) != 1 {
				t.Fatal("consumer acknowledged or advanced past an unapplied event")
			}
			if mode == "invalid envelope" && calls != 0 {
				t.Fatal("invalid envelope reached storage")
			}
		})
	}
}
