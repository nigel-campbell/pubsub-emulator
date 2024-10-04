package pstest

import (
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"os"
	"testing"
	"time"
)

func TestPersistentServer(t *testing.T) {
	// Remove the database file if it exists.
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	t.Logf("Current working directory: %s", cwd)
	_ = os.RemoveAll(fmt.Sprintf("%s/pubsub.db/", cwd))

	srv, err := NewServerWithCallback(3000, "pubsub.db", func(s *grpc.Server) {})
	if err != nil {
		t.Fatalf("failed to start Cloud Pub/Sub emulator: %v", err)
	}
	defer srv.Close()
	t.Logf("Starting Cloud Pub/Sub emulator on port %s", srv.Addr)

	// This Pub/Sub client is intended to run against an emulator in local dev.
	err = os.Setenv("PUBSUB_EMULATOR_HOST", fmt.Sprintf("127.0.0.1:%d", 3000))
	Ok(t, err, "failed to set emulator host envvar")

	ctx := context.Background()

	client, err := pubsub.NewClient(ctx, "dev", option.WithoutAuthentication())
	Ok(t, err, "failed to initialize Pub/Sub client")
	defer client.Close()

	_topic, err := client.CreateTopic(ctx, "new-topic")
	Ok(t, err, "failed to create topic")

	_subscription, err := client.CreateSubscription(ctx, "my-subscription", pubsub.SubscriptionConfig{
		Topic: _topic,
	})

	res := _topic.Publish(ctx, &pubsub.Message{
		Data: []byte("hello"),
	})
	_, err = res.Get(ctx)
	Ok(t, err, "failed to publish message")

	var msgs []*pubsub.Message
	ackHandler := func(ctx context.Context, msg *pubsub.Message) {
		msgs = append(msgs, msg)
		msg.Ack()
	}

	cctx, cancel := context.WithDeadline(ctx, time.Now().Add(3*time.Second))
	defer cancel()

	_subscription.ReceiveSettings.Synchronous = true
	err = _subscription.Receive(cctx, ackHandler)

	Ok(t, err, "failed to receive message")
	if len(msgs) != 1 {
		t.Fatalf("expected only 1 message, got %d", len(msgs))
	}
	t.Logf("recieved %d messages", len(msgs))

	res = _topic.Publish(ctx, &pubsub.Message{Data: []byte("hello")})
	_, err = res.Get(ctx)
	Ok(t, err, "failed to publish message")

	cctx, cancel = context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	err = _subscription.Receive(cctx, ackHandler)

	Ok(t, err, "failed to receive message")
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
}

func TestPersistentSubscriptionServer(t *testing.T) {
	// Remove the database file if it exists.
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	t.Logf("Current working directory: %s", cwd)
	_ = os.RemoveAll(fmt.Sprintf("%s/pubsub.db/", cwd))

	srv, err := NewServerWithCallback(3000, "pubsub.db", func(s *grpc.Server) {})
	if err != nil {
		t.Fatalf("failed to start Cloud Pub/Sub emulator: %v", err)
	}
	defer srv.Close()
	t.Logf("Starting Cloud Pub/Sub emulator on port %s", srv.Addr)

	// This Pub/Sub client is intended to run against an emulator in local dev.
	err = os.Setenv("PUBSUB_EMULATOR_HOST", fmt.Sprintf("127.0.0.1:%d", 3000))
	Ok(t, err, "failed to set emulator host envvar")

	ctx := context.Background()

	client, err := pubsub.NewClient(ctx, "dev", option.WithoutAuthentication())
	Ok(t, err, "failed to initialize Pub/Sub client")
	defer client.Close()

	_, err = client.CreateTopic(ctx, "new-topic")
	Ok(t, err, "failed to create topic")

	// TODO(nigel): Finish me!
}

func Ok(t *testing.T, err error, msg string) {
	if err != nil {
		t.Fatalf("%s: %v", msg, err)
	}
}
