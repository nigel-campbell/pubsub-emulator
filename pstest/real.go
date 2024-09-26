package pstest

import (
	pb "cloud.google.com/go/pubsub/apiv1/pubsubpb"
	"context"
	"errors"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	"sync"
	"time"
)

// Compile-time checks to ensure InmemGserver implements all interfaces
var _ pb.PublisherServer = (*PersistentGserver)(nil)
var _ pb.SubscriberServer = (*PersistentGserver)(nil)
var _ pb.SchemaServiceServer = (*PersistentGserver)(nil)
var _ SubscriptionServer = (*PersistentGserver)(nil)

type PersistentGserver struct {
	pb.UnimplementedPublisherServer
	pb.UnimplementedSchemaServiceServer
	UnimplementedSubscriptionServer

	// NB: LevelDB does not support transactions, so we need to lock around writes.
	mu      sync.Mutex
	db      *leveldb.DB
	nowFunc func() time.Time
}

func (u *PersistentGserver) SetTimeNowFunc(f func() time.Time) {
	u.nowFunc = f
}

func (p *PersistentGserver) CreateSubscription(ctx context.Context, s *pb.Subscription) (*pb.Subscription, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if the subscription already exists
	_, err := p.db.Get([]byte(s.Name), nil)
	if err == nil {
		// Subscription already exists
		return nil, fmt.Errorf("subscription %s already exists", s.Name)
	}

	if !errors.Is(err, leveldb.ErrNotFound) {
		// If there was an error other than "not found", return it
		return nil, err
	}

	// Marshal the subscription
	b, err := proto.Marshal(s)
	if err != nil {
		return nil, err
	}

	// Store the subscription in the database
	if err := p.db.Put([]byte(s.Name), b, nil); err != nil {
		return nil, err
	}

	return s, nil
}

func (p *PersistentGserver) GetSubscription(ctx context.Context, subscriptionRequest *pb.GetSubscriptionRequest) (*pb.Subscription, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PersistentGserver) UpdateSubscription(ctx context.Context, subscriptionRequest *pb.UpdateSubscriptionRequest) (*pb.Subscription, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PersistentGserver) ListSubscriptions(ctx context.Context, subscriptionsRequest *pb.ListSubscriptionsRequest) (*pb.ListSubscriptionsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PersistentGserver) DeleteSubscription(ctx context.Context, subscriptionRequest *pb.DeleteSubscriptionRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PersistentGserver) ModifyAckDeadline(ctx context.Context, deadlineRequest *pb.ModifyAckDeadlineRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PersistentGserver) Acknowledge(ctx context.Context, acknowledgeRequest *pb.AcknowledgeRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PersistentGserver) Pull(ctx context.Context, pullRequest *pb.PullRequest) (*pb.PullResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PersistentGserver) StreamingPull(server pb.Subscriber_StreamingPullServer) error {
	//TODO implement me
	panic("implement me")
}

func (p *PersistentGserver) ModifyPushConfig(ctx context.Context, configRequest *pb.ModifyPushConfigRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PersistentGserver) GetSnapshot(ctx context.Context, snapshotRequest *pb.GetSnapshotRequest) (*pb.Snapshot, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PersistentGserver) ListSnapshots(ctx context.Context, snapshotsRequest *pb.ListSnapshotsRequest) (*pb.ListSnapshotsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PersistentGserver) CreateSnapshot(ctx context.Context, snapshotRequest *pb.CreateSnapshotRequest) (*pb.Snapshot, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PersistentGserver) UpdateSnapshot(ctx context.Context, snapshotRequest *pb.UpdateSnapshotRequest) (*pb.Snapshot, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PersistentGserver) DeleteSnapshot(ctx context.Context, snapshotRequest *pb.DeleteSnapshotRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PersistentGserver) Seek(ctx context.Context, seekRequest *pb.SeekRequest) (*pb.SeekResponse, error) {
	//TODO implement me
	panic("implement me")
}

func topicRowKey(topic string) []byte {
	return []byte(fmt.Sprintf("topic:%s", topic))
}

func (p *PersistentGserver) CreateTopic(_ context.Context, t *pb.Topic) (*pb.Topic, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	// Check if the topic already exists
	_, err := p.db.Get(topicRowKey(t.Name), nil)
	if err == nil {
		// Topic already exists
		return nil, fmt.Errorf("topic %s already exists", t.Name)
	}

	if !errors.Is(err, leveldb.ErrNotFound) {
		// If there was an error other than "not found", return it
		return nil, err
	}

	b, err := proto.Marshal(t)
	if err != nil {
		return nil, err
	}
	if err := p.db.Put(topicRowKey(t.Name), b, nil); err != nil {
		return nil, err
	}
	return t, nil
}

func (p *PersistentGserver) UpdateTopic(ctx context.Context, req *pb.UpdateTopicRequest) (*pb.Topic, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if the topic exists
	existingTopicData, err := p.db.Get(topicRowKey(req.Topic.Name), nil)
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return nil, fmt.Errorf("topic %s not found", req.Topic.Name)
		}
		return nil, err
	}

	// Unmarshal the existing topic
	existingTopic := &pb.Topic{}
	if err := proto.Unmarshal(existingTopicData, existingTopic); err != nil {
		return nil, err
	}

	// Update the topic with new data
	updatedTopic := req.Topic

	// Marshal the updated topic
	updatedTopicData, err := proto.Marshal(updatedTopic)
	if err != nil {
		return nil, err
	}

	// Put the updated topic back into the database
	if err := p.db.Put([]byte(updatedTopic.Name), updatedTopicData, nil); err != nil {
		return nil, err
	}

	return updatedTopic, nil
}

func (p *PersistentGserver) Publish(_ context.Context, req *pb.PublishRequest) (*pb.PublishResponse, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if the topic exists
	topicData, err := p.db.Get(topicRowKey(req.Topic), nil)
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return nil, fmt.Errorf("topic %s not found", req.Topic)
		}
		return nil, err
	}

	// Unmarshal the existing topic
	topic := &pb.Topic{}
	if err := proto.Unmarshal(topicData, topic); err != nil {
		return nil, err
	}

	// Create a response
	response := &pb.PublishResponse{
		MessageIds: make([]string, len(req.Messages)),
	}

	// Iterate over the messages and store them
	for i, msg := range req.Messages {
		msgId := messageRowKey(req.Topic, msg.MessageId)
		msgData, err := proto.Marshal(msg)
		if err != nil {
			return nil, err
		}
		if err := p.db.Put(msgId, msgData, nil); err != nil {
			return nil, err
		}
		response.MessageIds[i] = msg.MessageId
	}
	return response, nil
}

func messageRowKey(topic string, msgId string) []byte {
	return []byte(fmt.Sprintf("#topic:%s#message:%s", topic, msgId))
}

func (p *PersistentGserver) GetTopic(ctx context.Context, req *pb.GetTopicRequest) (*pb.Topic, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PersistentGserver) ListTopics(ctx context.Context, topicsRequest *pb.ListTopicsRequest) (*pb.ListTopicsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PersistentGserver) ListTopicSubscriptions(ctx context.Context, subscriptionsRequest *pb.ListTopicSubscriptionsRequest) (*pb.ListTopicSubscriptionsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PersistentGserver) ListTopicSnapshots(ctx context.Context, snapshotsRequest *pb.ListTopicSnapshotsRequest) (*pb.ListTopicSnapshotsResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PersistentGserver) DeleteTopic(ctx context.Context, topicRequest *pb.DeleteTopicRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PersistentGserver) DetachSubscription(ctx context.Context, subscriptionRequest *pb.DetachSubscriptionRequest) (*pb.DetachSubscriptionResponse, error) {
	//TODO implement me
	panic("implement me")
}

// UnimplementedSubscriptionServer is a mock implementation of SubscriptionServer which panics on all methods
type UnimplementedSubscriptionServer struct{}

var _ SubscriptionServer = (*UnimplementedSubscriptionServer)(nil)

func (u UnimplementedSubscriptionServer) Messages() []*Message {
	//TODO implement me
	panic("implement me")
}

func (u UnimplementedSubscriptionServer) Message(id string) *Message {
	//TODO implement me
	panic("implement me")
}

func (u UnimplementedSubscriptionServer) Close() error {
	//TODO implement me
	panic("implement me")
}

func (u UnimplementedSubscriptionServer) Wait() {
	//TODO implement me
	panic("implement me")
}

func (u UnimplementedSubscriptionServer) SetStreamTimeout(d time.Duration) {
	//TODO implement me
	panic("implement me")
}

func (u UnimplementedSubscriptionServer) ResetPublishResponses(size int) {
	//TODO implement me
	panic("implement me")
}

func (u UnimplementedSubscriptionServer) SetTimeNowFunc(f func() time.Time) {
	//TODO implement me
	panic("implement me")
}

func (u UnimplementedSubscriptionServer) SetAutoPublishResponse(rsp bool) {
	//TODO implement me
	panic("implement me")
}

func (u UnimplementedSubscriptionServer) AddPublishResponse(pbr *pb.PublishResponse, err error) {
	//TODO implement me
	panic("implement me")
}

func (u UnimplementedSubscriptionServer) Topics() map[string]*topic {
	//TODO implement me
	panic("implement me")
}

func (u UnimplementedSubscriptionServer) Subscriptions() map[string]*subscription {
	//TODO implement me
	panic("implement me")
}

func (u UnimplementedSubscriptionServer) Schemas() map[string][]*pb.Schema {
	//TODO implement me
	panic("implement me")
}

func (u UnimplementedSubscriptionServer) ClearMessages() {
	//TODO implement me
	panic("implement me")
}