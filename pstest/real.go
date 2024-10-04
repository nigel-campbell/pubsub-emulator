package pstest

import (
	"bytes"
	pb "cloud.google.com/go/pubsub/apiv1/pubsubpb"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"math/rand"
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
	if err := p.db.Put(subscriptionRowKey(s.Topic, s.Name), b, nil); err != nil {
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

func (p *PersistentGserver) ModifyAckDeadline(ctx context.Context, req *pb.ModifyAckDeadlineRequest) (*emptypb.Empty, error) {
	// TODO implement me but don't panic! Messages are acked by deleting them from the database for now.
	return &emptypb.Empty{}, nil
}

func (p *PersistentGserver) Acknowledge(ctx context.Context, req *pb.AcknowledgeRequest) (*emptypb.Empty, error) {
	// If the message is ackd, simply delete the message. This is pretty poor way to implement this but I want to
	// get SOMETHING done and this achieves the desired affect of preventing the message from being redelivered so let's
	// try it. Also, for now the ack id will match the message id.
	for _, ackId := range req.AckIds {
		// delete messages matching the ack id
		err := p.db.Delete([]byte(ackId), nil)
		if err != nil {
			log.Println("failed to delete message with ack id", ackId)
		}
	}
	return &emptypb.Empty{}, nil
}

// DeliveredMessage is a struct that represents a message that has been delivered to a subscriber
// It stores the message id and the delivery timestamp to enforce the ack deadline.
type DeliveredMessage struct {
	MessageID         string
	DeliveryTimestamp int64
}

func (p *PersistentGserver) Pull(_ context.Context, req *pb.PullRequest) (*pb.PullResponse, error) {
	defer p.mu.Unlock()
	p.mu.Lock()

	pulledMessages, err := getMessages(p.db, req.Subscription)
	if err != nil {
		return nil, err
	}

	// fetch all acks for the subscription
	deliveredMsgs, err := readAllAcks(p.db, req.Subscription)
	if err != nil {
		return nil, err
	}

	validMessages := pulledMessages
	if len(deliveredMsgs) > 0 {
		validMessages = filterMessages(pulledMessages, deliveredMsgs)
	}
	//log.Printf("pulled %d messages, %d valid messages", len(pulledMessages), len(validMessages))

	if err := writeAcks(p.db, req.Subscription, validMessages); err != nil {
		return nil, err
	}

	return &pb.PullResponse{
		ReceivedMessages: validMessages,
	}, nil
}

// filterMessages filters out messages that have already been delivered and past their acknowledgement deadline
func filterMessages(rcvdMsgs []*pb.ReceivedMessage, deliveredMsgs map[string]int64) []*pb.ReceivedMessage {
	var validMessages []*pb.ReceivedMessage
	for _, msg := range rcvdMsgs {
		deliveryTime, ok := deliveredMsgs[msg.Message.MessageId]
		timeUntilDeadline := time.Since(time.Unix(deliveryTime, 0))
		if ok && timeUntilDeadline < time.Second*30 { // TODO(nigel): This should be configurable per subscriptpion
			continue
		}
		log.Printf("current time is %v, message delivered is %v (%s until deadline)", time.Now(), deliveryTime, timeUntilDeadline)
		validMessages = append(validMessages, msg)
	}
	return validMessages
}

func ackIds(rcvdMsgs []*pb.ReceivedMessage) []string {
	acks := make([]string, len(rcvdMsgs))
	for i, rcvdMsg := range rcvdMsgs {
		acks[i] = rcvdMsg.AckId
	}
	return acks
}

func readAcks(db *leveldb.DB, subscriptionId string, ackIds []string) (map[string]int64, error) {
	msgToTimestamp := make(map[string]int64) // messageId -> deliveryTimestamp
	for _, ackId := range ackIds {
		val, err := db.Get(ackRowKey(subscriptionId, ackId), nil)
		if err != nil && !errors.Is(err, leveldb.ErrNotFound) {
			return nil, err
		}
		if errors.Is(err, leveldb.ErrNotFound) {
			continue
		}
		var deliveredMessage DeliveredMessage
		if err := gob.NewDecoder(bytes.NewReader(val)).Decode(&deliveredMessage); err != nil {
			return nil, err
		}
		msgToTimestamp[deliveredMessage.MessageID] = deliveredMessage.DeliveryTimestamp
		log.Printf("read ack for message %s with delivery timestamp %v", deliveredMessage.MessageID, time.Unix(deliveredMessage.DeliveryTimestamp, 0))
	}
	return msgToTimestamp, nil
}

func readAllAcks(db *leveldb.DB, subscriptionId string) (map[string]int64, error) {
	prefix := []byte(fmt.Sprintf("#subscription:%s#ack:", subscriptionId))
	iter := db.NewIterator(util.BytesPrefix(prefix), nil)
	defer iter.Release()
	msgToTimestamp := make(map[string]int64)
	for iter.Next() {
		var deliveredMessage DeliveredMessage
		if err := gob.NewDecoder(bytes.NewReader(iter.Value())).Decode(&deliveredMessage); err != nil {
			return nil, err
		}
		msgToTimestamp[deliveredMessage.MessageID] = deliveredMessage.DeliveryTimestamp
	}
	return msgToTimestamp, nil
}

func writeAcks(db *leveldb.DB, subscriptionId string, rcvdMsgs []*pb.ReceivedMessage) error {
	deliveryTimestamp := time.Now().Unix()
	for _, rcvdMsg := range rcvdMsgs {
		msg := DeliveredMessage{
			MessageID:         rcvdMsg.Message.MessageId,
			DeliveryTimestamp: deliveryTimestamp,
		}
		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(&msg); err != nil {
			return err // This should never fail because we're encoding a struct
		}
		err := db.Put(ackRowKey(subscriptionId, rcvdMsg.AckId), buf.Bytes(), nil)
		if err != nil {
			return err
		}
		log.Printf("wrote ack id %s for message %s with delivery timestamp %v", rcvdMsg.AckId, msg.MessageID, time.Unix(msg.DeliveryTimestamp, 0))
	}
	return nil
}

func getMessages(db *leveldb.DB, subscriptionId string) ([]*pb.ReceivedMessage, error) {
	prefix := []byte(fmt.Sprintf("#subscription:%s#message:", subscriptionId))
	iter := db.NewIterator(util.BytesPrefix(prefix), nil)
	defer iter.Release()
	var rcvdMsgs []*pb.ReceivedMessage
	for iter.Next() {
		var msg pb.PubsubMessage
		val := iter.Value()
		err := proto.Unmarshal(val, &msg)
		if err != nil {
			return nil, err
		}

		// the ackId is just a random string for now
		ackId := fmt.Sprintf("%d", rand.Int())
		rcvdMsgs = append(rcvdMsgs, &pb.ReceivedMessage{
			Message: &msg,
			AckId:   ackId,
		})
	}
	return rcvdMsgs, nil
}

func ackRowKey(subscriptionId, ackId string) []byte {
	return []byte(fmt.Sprintf("#subscription:%s#ack:%s", subscriptionId, ackId))
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
	t := &pb.Topic{}
	if err := proto.Unmarshal(topicData, t); err != nil {
		return nil, err
	}

	response := &pb.PublishResponse{
		MessageIds: make([]string, len(req.Messages)),
	}

	// Fetch all subscriptions given the topic
	var subscriptions []string

	prefix := topicRowKey(req.Topic)
	// Use an iterator to scan for the prefix
	iter := p.db.NewIterator(util.BytesPrefix(prefix), nil)
	defer iter.Release()
	for iter.Next() {
		if string(iter.Key()) != string(prefix) {
			_, sub, err := parseSubscriptionRowKey(iter.Key())
			if err != nil {
				return nil, fmt.Errorf("failed to parse subscription row key")
			}
			subscriptions = append(subscriptions, sub)
		}
	}

	for _, sub := range subscriptions {
		for i, msg := range req.Messages {
			msg.MessageId = fmt.Sprintf("%d", rand.Int())
			msgId := messageRowKey(sub, msg.MessageId)
			msgData, err := proto.Marshal(msg)
			if err != nil {
				return nil, err
			}
			if err := p.db.Put(msgId, msgData, nil); err != nil {
				return nil, err
			}
			response.MessageIds[i] = msg.MessageId
		}
	}
	return response, nil
}

func messageRowKey(subscription, msgId string) []byte {
	return []byte(fmt.Sprintf("#subscription:%s#message:%s", subscription, msgId))
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

func (u UnimplementedSubscriptionServer) Close() error { return nil }

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
