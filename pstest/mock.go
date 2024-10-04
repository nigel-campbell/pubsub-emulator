package pstest

import (
	pb "cloud.google.com/go/pubsub/apiv1/pubsubpb"
	"time"
)

// UnimplementedInmemoryServer is a mock implementation of InmemServer which panics on all methods
// It's sole purpose is to ensure that the InmemServer interface matches the PersistentServer interface
// TODO(nigel): Remove this in favor of implementing actual methods on the PersistentServer
type UnimplementedInmemoryServer struct{}

var _ InmemoryServer = (*UnimplementedInmemoryServer)(nil)

func (u UnimplementedInmemoryServer) Messages() []*Message {
	//TODO implement me
	panic("implement me")
}

func (u UnimplementedInmemoryServer) Message(id string) *Message {
	//TODO implement me
	panic("implement me")
}

func (u UnimplementedInmemoryServer) Close() error { return nil }

func (u UnimplementedInmemoryServer) Wait() {
	//TODO implement me
	panic("implement me")
}

func (u UnimplementedInmemoryServer) SetStreamTimeout(d time.Duration) {
	//TODO implement me
	panic("implement me")
}

func (u UnimplementedInmemoryServer) ResetPublishResponses(size int) {
	//TODO implement me
	panic("implement me")
}

func (u UnimplementedInmemoryServer) SetTimeNowFunc(f func() time.Time) {
	//TODO implement me
	panic("implement me")
}

func (u UnimplementedInmemoryServer) SetAutoPublishResponse(rsp bool) {
	//TODO implement me
	panic("implement me")
}

func (u UnimplementedInmemoryServer) AddPublishResponse(pbr *pb.PublishResponse, err error) {
	//TODO implement me
	panic("implement me")
}

func (u UnimplementedInmemoryServer) Topics() map[string]*topic {
	//TODO implement me
	panic("implement me")
}

func (u UnimplementedInmemoryServer) Subscriptions() map[string]*subscription {
	//TODO implement me
	panic("implement me")
}

func (u UnimplementedInmemoryServer) Schemas() map[string][]*pb.Schema {
	//TODO implement me
	panic("implement me")
}

func (u UnimplementedInmemoryServer) ClearMessages() {
	//TODO implement me
	panic("implement me")
}
