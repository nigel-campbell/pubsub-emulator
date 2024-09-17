package main

import (
	"context"
	"errors"
	"flag"
	"google.golang.org/grpc"
	"log"
	"pubsub-emulator/pstest"
)

var (
	host = flag.String("host", "localhost", "the address to bind to on the local machine")
	port = flag.Int("port", 8085, "the port number to bind to on the local machine")
	dir  = flag.String("dir", "", "if set, use persistence in the given directory")
)

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	grpc.EnableTracing = true
	flag.Parse()

	if *dir != "" {
		// TODO(nigel): Implement persistence
		return errors.New("not yet implemented")
	}

	// TODO(nigel): Add grpcui for quick admin page
	srv := pstest.NewServerWithPort(*port)
	defer srv.Close()
	log.Printf("Starting Cloud Pub/Sub emulator on port %s", srv.Addr)

	// TODO(nigel): Ensure pstest respects context deadline and SIGTERM
	select {
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}
