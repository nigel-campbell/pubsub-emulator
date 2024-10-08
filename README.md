# Pub/Sub Emulator

This is an "out of process" emulator built on top of the emulator 
provided by https://github.com/googleapis/google-cloud-go/tree/main/pubsub/pstest.
The difference is that this package intends to provide a persistence option.

## Running

This tool can be built using the `go` tool.

```sh
go build pubsub-emulator/cmd/... && ./pubsub-emulator
```

The driver tool can be built using the `go` tool as well.

```sh
go build pubsub-emulator/example/client/... && ./client
```

An admin page is available on `localhost:8080`.

## Goals

- Topics, subscriptions and published messages should persist after the emulator shuts down.
- An admin page for viewing topics, subscriptions and messages

## Non-Goals

- Durability beyond what's feasible with an embedded database alone.
- Validation, testing against the real Pub/Sub API

## TODOs

   - [x] Setup repo and fork copy of [pstest package](https://github.com/googleapis/google-cloud-go/tree/main/pubsub/pstest) (see example [here](https://github.com/fullstorydev/emulators))
   - [x] Stub out in-memory Pub/Sub emulator binary
   - [x] Stub out driver to validate:
     - [x] Topic and subscription creation/modification/destruction
     - [x] Handle message publishing
     - [x] Subscription pulling and message acknolwedgements
   - [x] Admin page for viewing topics, subscriptions and messages
   - [ ] Add Makefile or build script to simplify build process
   - [x] Add persistence layer (see PersistentGServer in real.go and milestones below)
   - [ ] Ensure existing unit tests cover both in-memory and persistent modes

## Milestones

  - [x] Compiles and builds in local dev (ideally with CI)
  - [x] Supports in-memory topics/subscriptions 
  - [x] Driver tool for publishing and subscribing to messages in local dev
  - [x] Supports persistence
    - [x] Messages published to a given topic persist after emulator restart
    - [x] Messages delivery and acknowledgement is supported
    - [x] Ack deadline can be modified (to support NACKs)
    - [ ] StreamingPull is supported
  - [x] UI for inspecting/debugging emulator contents

## Constraints

Must be self-contained and a single binary (i.e. use an embedded database rather than Kafka, Postgres or RabbitMQ)

## LevelDB Schema

Messages are stored using the following key format `#subscription:<subscription-id>#message:<message-id>`.

Topics are stored using the following key format `#topic:<topic-name>`.

Subscriptions are stored using the following key format `#topic:<topic-name>#subscription:<subscription-name>`.

Acknowledgements are stored using the following key format `#subscription:<subscription-id>#ack:<message-id>`.