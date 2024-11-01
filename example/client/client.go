// Command client is a Pub/Sub client that demonstrates how to use the Pub/Sub client library with the Pub/Sub emulator.
package main

import (
	"cloud.google.com/go/pubsub"
	"context"
	"errors"
	"flag"
	"fmt"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"math/rand"
	"os"
	"time"
)

const (
	devProjectId     = "dev" // NB: This value doesn't actually matter to the Pub/Sub client when initializing against the emulator
	topicName        = "my-topic"
	subscriptionName = "my-subscription"
	workerCount      = 3
)

var (
	port          = flag.Int("port", 8085, "port of the Pub/Sub emulator")
	synchronous   = flag.Bool("synchronous", true, "use synchronous pull (when false, uses streaming pull)")
	pull          = flag.Bool("pull", true, "whether or not the client pulls")
	nackRate      = flag.Float64("nack-rate", 0.0, "rate at which to nack messages")
	shouldPublish = flag.Bool("publish", true, "whether or not the client publishes")
)

func main() {
	if err := run(context.Background()); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("exiting prematurely with: %v", err)
	}
}

func run(ctx context.Context) error {
	flag.Parse()

	// This Pub/Sub client is intended to run against an emulator in local dev.
	if err := os.Setenv("PUBSUB_EMULATOR_HOST", fmt.Sprintf("127.0.0.1:%d", *port)); err != nil {
		return fmt.Errorf("failed to set emulator host envvar: %w", err)
	}

	c, err := pubsub.NewClient(ctx, devProjectId, option.WithoutAuthentication())
	if err != nil {
		return fmt.Errorf("failed to initialize Pub/Sub client: %w", err)
	}
	defer c.Close()

	log.Printf("initializing new topic against %s\n", os.Getenv("PUBSUB_EMULATOR_HOST"))
	topic, err := c.CreateTopic(ctx, topicName)
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			topic = c.Topic(topicName)
		} else {
			return fmt.Errorf("failed to create topic: %w", err)
		}
	}
	log.Println("new topic created")

	subscription, err := c.CreateSubscription(ctx, subscriptionName, pubsub.SubscriptionConfig{
		Topic: topic,
	})
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			subscription = c.Subscription(subscriptionName)
		} else {
			return fmt.Errorf("failed to create subscription: %w", err)
		}
	}
	log.Println("new subscription created")

	done := make(chan struct{})

	// Receive messages concurrently.
	if *pull {
		go func() {
			defer func() {
				close(done) // In case the subscription exits early.
			}()
			subscription.ReceiveSettings.Synchronous = *synchronous
			err := subscription.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
				log.Printf("got message: %q (messageId=%s)\n", string(msg.Data), msg.ID)
				if shouldNack() {
					log.Printf("nacking message: %q\n", string(msg.Data))
					msg.Nack()
					return
				}
				msg.Ack()
			})
			if err != nil {
				log.Printf("failed to receive message: %v", err)
			}
		}()
	}

	// Spawn multiple goroutines to publish messages until the context is done.
	grp, ctx := errgroup.WithContext(ctx)
	for i := 0; i < workerCount; i++ {
		i := i
		grp.Go(func() error {
			return publishMessages(ctx, i, topic, done)
		})
	}

	// Setup subscription to receive messages.
	if err := grp.Wait(); err != nil {
		return fmt.Errorf("failed to publish messages: %w", err)
	}
	return nil
}

func shouldNack() bool {
	if *nackRate >= 1 {
		return true
	}
	if *nackRate > 0 && *nackRate < 1 {
		return rand.Float64() < *nackRate
	}
	return false
}

func publishMessages(ctx context.Context, workerId int, topic *pubsub.Topic, done chan struct{}) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-done:
			return nil
		default:
			if *shouldPublish {
				data := []byte(fmt.Sprintf("It's %s from worker %d", time.Now().Format(time.RFC3339), workerId))
				res := topic.Publish(ctx, &pubsub.Message{
					Data: data,
				})
				if _, err := res.Get(ctx); err != nil {
					return fmt.Errorf("failed to publish message: %w", err)
				}
				// Sleep for a random amount of time between 0 and 10 seconds.
				time.Sleep(time.Second * time.Duration(time.Now().UnixNano()%10))
				if !*pull {
					log.Printf("published message: %q\n", string(data))
				}
			}
		}
	}
	return nil
}
