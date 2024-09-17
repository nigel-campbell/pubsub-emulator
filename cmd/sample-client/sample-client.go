// Command sample-client is a Pub/Sub client that demonstrates how to use the Pub/Sub client library with the Pub/Sub emulator.
package main

import (
	"cloud.google.com/go/pubsub"
	"context"
	"flag"
	"fmt"
	"google.golang.org/api/option"
	"log"
	"os"
	"time"
)

const (
	devProjectId = "dev" // NB: This value doesn't actually matter to the Pub/Sub client when initializing against the emulator
	topicName    = "my-topic"
)

var (
	port = flag.Int("port", 8085, "port of the Pub/Sub emulator")
)

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	flag.Parse()
	// This Pub/Sub client is intended to run against an emulator in local dev.
	// Ensure that the emulator host envvar is set appropriately.
	// See https://cloud.google.com/pubsub/docs/emulator for more information.
	if err := os.Setenv("PUBSUB_EMULATOR_HOST", fmt.Sprintf("127.0.0.1:%d", *port)); err != nil {
		return fmt.Errorf("failed to set emulator host envvar: %w", err)
	}

	c, err := pubsub.NewClient(ctx, devProjectId, option.WithoutAuthentication())
	if err != nil {
		return fmt.Errorf("failed to initialize Pub/Sub client: %w", err)
	}
	defer c.Close()

	deadline := time.Now().Add(15 * time.Second)
	log.Printf("initializing new topic against %s\n", os.Getenv("PUBSUB_EMULATOR_HOST"))

	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()
	if _, err = c.CreateTopic(ctx, topicName); err != nil {
		return fmt.Errorf("failed to create topic: %w", err)
	}
	log.Println("new topic created")
	return nil
}
