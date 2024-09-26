package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/fullstorydev/grpcui/standalone"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net/http"
	"pubsub-emulator/pstest"
)

var (
	host = flag.String("host", "localhost", "the address to bind to on the local machine")
	port = flag.Int("port", 8085, "the port number to bind to on the local machine")
	dir  = flag.String("dir", "", "if set, use dbclient in the given directory")
)

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	grpc.EnableTracing = true
	flag.Parse()

	srv, err := pstest.NewServerWithCallback(*port, *dir, func(s *grpc.Server) {})
	if err != nil {
		return fmt.Errorf("failed to start Cloud Pub/Sub emulator: %w", err)
	}
	defer srv.Close()
	log.Printf("Starting Cloud Pub/Sub emulator on port %s", srv.Addr)

	startGrpcui(ctx)

	// TODO(nigel): Ensure pstest respects SIGTERM
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-srv.Done():
		log.Printf("Cloud Pub/Sub emulator on port %s has stopped", srv.Addr)
		return nil
	}
}

func startGrpcui(ctx context.Context) {
	// Start HTTP server and integrate grpcui
	go func() {
		// Create gRPC-UI handler
		httpMux := http.NewServeMux()

		// Dial gRPC server running locally
		target := fmt.Sprintf("localhost:%d", *port)
		grpcConn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Failed to dial gRPC server: %v", err)
		}

		// Setup grpcui handler
		grpcUIHandler, err := standalone.HandlerViaReflection(ctx, grpcConn, target)
		if err != nil {
			log.Fatalf("Failed to create grpcui handler: %v", err)
		}

		// Mount the grpcui handler at a path, e.g., /grpcui
		httpMux.Handle("/grpcui/", http.StripPrefix("/grpcui", grpcUIHandler))

		log.Println("Starting HTTP server for grpcui on :8080")
		if err := http.ListenAndServe(":8080", httpMux); err != nil {
			log.Fatalf("Failed to serve HTTP: %v", err)
		}
	}()
	select {}
}
