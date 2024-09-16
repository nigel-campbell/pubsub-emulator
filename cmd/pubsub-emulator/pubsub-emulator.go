package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"pubsub/testutil"
)

var (
	host = flag.String("host", "localhost", "the address to bind to on the local machine")
	port = flag.Int("port", 9000, "the port number to bind to on the local machine")
	dir  = flag.String("dir", "", "if set, use persistence in the given directory")
)

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(_ context.Context) error {
	fmt.Println("Starting Cloud Pub/Sub emulator")

	grpc.EnableTracing = true
	flag.Parse()

	if *dir != "" {
		// TODO(nigel): Implement persistence
		return errors.New("not yet implemented")
	}

	// TODO(nigel): Add grpcui to ensure that
	srv, err := testutil.NewServerWithPort(*port)
	if err != nil {
		return err
	}

	// TODO(nigel): Ensure that the context is respected
	return srv.BlockingStart()
}
