package dbclient

import "fmt"

type Topic struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Subscription struct {
	ID       int    `json:"id"`
	TopicID  int    `json:"topic_id"`
	Endpoint string `json:"endpoint"`
}

type Message struct {
	ID             int    `json:"id"`
	TopicID        int    `json:"topic_id"`
	SubscriptionID int    `json:"subscription_id"`
	Content        string `json:"content"`
}

// TODO(nigel): Remove the row key madness if we decide to go with a real database.
func (t *Topic) GetRowKey() string {
	return fmt.Sprintf("topic:%d", t.ID)
}

func (s *Subscription) GetRowKey() string {
	return fmt.Sprintf("topic:%d:subscription:%d", s.TopicID, s.ID)
}

func (s *Message) GetRowKey() string {
	return fmt.Sprintf("topic:%d:subscription:%d:message:%d", s.TopicID, s.SubscriptionID, s.ID)
}
