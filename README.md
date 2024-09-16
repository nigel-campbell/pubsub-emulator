# Pub/Sub Emulator

This is an "out of process" emulator built on top of the emulator 
provided by https://github.com/googleapis/google-cloud-go/tree/main/pubsub/pstest.
The difference is that this package intends to provide a persistence option.

## Goal

Topics, subscriptions and published messages should persist after the emulator shuts down.

## Stretch goal

Add support for emulator inspection (i.e. see topics, subscriptions, messages) 

## TODOs

   - [x] Setup repo and fork copy of [pstest package](https://github.com/googleapis/google-cloud-go/tree/main/pubsub/pstest) (see example [here](https://github.com/fullstorydev/emulators))
   - [x] Stub out in-memory Pub/Sub emulator binary
   - [ ] Stub out driver to validate:
     - [ ] Topic and subscription creation/modification/destruction
     - [ ] Message publishing
     - [ ] Subscription pulling
   - [ ] Add Makefile or build script

## Milestones

  - [x] Compiles and builds in local dev (ideally with CI)
  - [ ] Supports in-memory topics/subscriptions 
    - [ ] Comes with pstest package but needs to be validated
  - [ ] Driver tool for publishing and subscribing to messages in local dev
  - [ ] Supports persistence
  - [ ] UI for inspecting/debugging emulator contents

## Constraints

Must be self-contained and a single binary (i.e. use an embedded database rather than Kafka, Postgres or RabbitMQ)