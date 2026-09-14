// Package kafkatransport carries ledger references; financial payloads stay in MongoDB.
package kafkatransport

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"smlcloudplatform/internal/config"
	gl "smlcloudplatform/internal/generalledger"
	"smlcloudplatform/internal/product/projection"
)

const Topic = "bc-gl-projection-v1"
const operationTimeout = 5 * time.Second
const maxReferenceBytes = 4096

type messageWriter interface {
	WriteMessages(context.Context, ...kafka.Message) error
	Close() error
}

type messageReader interface {
	projection.Reader
	Close() error
}

type Bus struct {
	writer    messageWriter
	transport *kafka.Transport
	newReader func() messageReader
}

var _ gl.EventPublisher = (*Bus)(nil)

func New(mq config.IMQConfig, group string) (*Bus, error) {
	return NewForTopic(mq, group, Topic)
}

// NewForTopic lets isolated integration tests provision their own topic.
// Topics must already exist; publishing never creates one implicitly.
func NewForTopic(mq config.IMQConfig, group, topic string) (*Bus, error) {
	if mq == nil || strings.TrimSpace(mq.URI()) == "" {
		return nil, errors.New("GL Kafka brokers are required")
	}
	if strings.TrimSpace(group) == "" || len(topic) > 249 || topic == "." || topic == ".." || !regexp.MustCompile(`^[A-Za-z0-9._-]+$`).MatchString(topic) {
		return nil, errors.New("GL Kafka group or topic is invalid")
	}
	brokers := strings.Split(mq.URI(), ",")
	for i, broker := range brokers {
		broker = strings.TrimSpace(broker)
		host, port, err := net.SplitHostPort(broker)
		n, portErr := strconv.Atoi(port)
		if err != nil || portErr != nil || n < 1 || n > 65535 || host == "" || strings.ContainsAny(host, "@/?# \t\r\n") {
			return nil, errors.New("GL Kafka broker must be a host and port")
		}
		brokers[i] = broker
	}
	tlsConfig, err := transportTLS(mq)
	if err != nil {
		return nil, err
	}
	transport := &kafka.Transport{DialTimeout: operationTimeout, TLS: tlsConfig}
	writer := &kafka.Writer{
		Addr: kafka.TCP(brokers...), Topic: topic, Balancer: &kafka.Hash{},
		RequiredAcks: kafka.RequireAll, Async: false, MaxAttempts: 1,
		BatchSize: 1, ReadTimeout: operationTimeout, WriteTimeout: operationTimeout,
		Transport: transport, AllowAutoTopicCreation: false,
	}
	readerConfig := kafka.ReaderConfig{
		Brokers: brokers, Topic: topic, GroupID: strings.TrimSpace(group),
		Dialer:      &kafka.Dialer{Timeout: operationTimeout, TLS: tlsConfig},
		StartOffset: kafka.FirstOffset, CommitInterval: 0,
		MinBytes: 1, MaxBytes: maxReferenceBytes, MaxWait: time.Second,
		ReadBatchTimeout: operationTimeout,
	}
	return &Bus{writer: writer, transport: transport, newReader: func() messageReader { return kafka.NewReader(readerConfig) }}, nil
}

func transportTLS(mq config.IMQConfig) (*tls.Config, error) {
	protocol := strings.ToUpper(strings.TrimSpace(mq.SecurityProtocol()))
	if protocol == "" || protocol == "PLAINTEXT" {
		if mq.SSLCAFile() != "" || mq.SSLCertFile() != "" || mq.SSLKeyFile() != "" {
			return nil, errors.New("GL Kafka TLS files require SSL protocol")
		}
		return nil, nil
	}
	if protocol != "SSL" {
		return nil, errors.New("GL Kafka supports only PLAINTEXT or SSL")
	}
	roots, err := x509.SystemCertPool()
	if err != nil || roots == nil {
		roots = x509.NewCertPool()
	}
	if mq.SSLCAFile() != "" {
		pem, err := os.ReadFile(mq.SSLCAFile())
		if err != nil || !roots.AppendCertsFromPEM(pem) {
			return nil, errors.New("GL Kafka TLS CA file is unreadable or invalid")
		}
	}
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12, RootCAs: roots}
	if (mq.SSLCertFile() == "") != (mq.SSLKeyFile() == "") {
		return nil, errors.New("GL Kafka TLS client certificate and key must be configured together")
	}
	if mq.SSLCertFile() != "" {
		certificate, err := tls.LoadX509KeyPair(mq.SSLCertFile(), mq.SSLKeyFile())
		if err != nil {
			return nil, errors.New("GL Kafka TLS client certificate or key is invalid")
		}
		tlsConfig.Certificates = []tls.Certificate{certificate}
	}
	return tlsConfig, nil
}

func (b *Bus) Publish(ctx context.Context, event gl.Event) error {
	ref, err := gl.ReferenceFor(event)
	if err != nil {
		return err
	}
	value, err := json.Marshal(ref)
	if err != nil {
		return err
	}
	message := kafka.Message{Key: []byte(gl.ProjectionKey(ref.HoldingCode, ref.BusinessCode)), Value: value}
	if _, err := DecodeReference(message); err != nil {
		return err
	}
	publishCtx, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()
	return b.writer.WriteMessages(publishCtx, message)
}

// DecodeReference rejects duplicate/unknown fields, trailing JSON and mismatched
// partition keys before the consumer loads the authoritative MongoDB event.
func DecodeReference(message kafka.Message) (gl.EventReference, error) {
	var ref gl.EventReference
	invalid := errors.New("invalid GL Kafka event reference")
	if len(message.Value) == 0 || len(message.Value) > maxReferenceBytes {
		return ref, invalid
	}
	d := json.NewDecoder(bytes.NewReader(message.Value))
	first, err := d.Token()
	if err != nil || first != json.Delim('{') {
		return ref, invalid
	}
	seen := map[string]bool{}
	for d.More() {
		token, err := d.Token()
		name, ok := token.(string)
		if err != nil || !ok || seen[name] {
			return ref, invalid
		}
		seen[name] = true
		var target interface{}
		switch name {
		case "schemaversion":
			target = &ref.SchemaVersion
		case "eventid":
			target = &ref.EventID
		case "holdingcode":
			target = &ref.HoldingCode
		case "businesscode":
			target = &ref.BusinessCode
		case "sequence":
			target = &ref.Sequence
		case "eventhash":
			target = &ref.EventHash
		default:
			return ref, invalid
		}
		if err := d.Decode(target); err != nil {
			return ref, invalid
		}
	}
	if token, err := d.Token(); err != nil || token != json.Delim('}') {
		return ref, invalid
	}
	if _, err := d.Token(); err != io.EOF {
		return ref, invalid
	}
	if len(seen) != 6 || ref.Validate() != nil || string(message.Key) != gl.ProjectionKey(ref.HoldingCode, ref.BusinessCode) {
		return ref, invalid
	}
	return ref, nil
}

type boundedReader struct{ messageReader }

func (r boundedReader) CommitMessages(ctx context.Context, messages ...kafka.Message) error {
	commitCtx, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()
	return r.messageReader.CommitMessages(commitCtx, messages...)
}

// Run closes the failed reader before retrying the same group. Financial errors
// never use projection.ErrRejected, so a failed event cannot advance its offset.
func (b *Bus) Run(ctx context.Context, apply func(context.Context, gl.EventReference) error, report func(error)) {
	for ctx.Err() == nil {
		reader := b.newReader()
		err := projection.Consume(ctx, boundedReader{reader}, func(message kafka.Message) (err error) {
			defer func() {
				if recover() != nil {
					err = errors.New("GL Kafka consumer handler panicked")
				}
			}()
			ref, err := DecodeReference(message)
			if err != nil {
				return err
			}
			if err = apply(ctx, ref); err != nil {
				// Do not inherit the product consumer's acknowledge-and-skip policy.
				return fmt.Errorf("GL projection not applied: %v", err)
			}
			return nil
		})
		_ = reader.Close()
		if ctx.Err() != nil {
			return
		}
		if err != nil && report != nil {
			report(err)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second):
		}
	}
}

func (b *Bus) Close() error {
	err := b.writer.Close()
	b.transport.CloseIdleConnections()
	return err
}
