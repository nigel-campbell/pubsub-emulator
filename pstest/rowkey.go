package pstest

import (
	"fmt"
	"regexp"
)

func topicRowKey(topic string) []byte {
	return []byte(fmt.Sprintf("#topic:%s", topic))
}

// Parser for topicRowKey
func parseTopicRowKey(rowKey []byte) (string, error) {
	// Convert byte slice to string
	rowKeyStr := string(rowKey)

	// Use regular expressions to extract topic
	re := regexp.MustCompile(`#topic:(.*)`)
	matches := re.FindStringSubmatch(rowKeyStr)

	if len(matches) < 2 {
		return "", fmt.Errorf("invalid topic row key format")
	}

	// Return the topic
	return matches[1], nil
}

func subscriptionRowKey(topic, subscription string) []byte {
	return []byte(fmt.Sprintf("#topic:%s#subscription:%s", topic, subscription))
}

// Parser for subscriptionRowKey
func parseSubscriptionRowKey(rowKey []byte) (string, string, error) {
	// Convert byte slice to string
	rowKeyStr := string(rowKey)

	// Use regular expressions to extract topic and subscription
	re := regexp.MustCompile(`#topic:(.*)#subscription:(.*)`)
	matches := re.FindStringSubmatch(rowKeyStr)

	if len(matches) < 3 {
		return "", "", fmt.Errorf("invalid subscription row key format")
	}

	// Return the topic and subscription
	return matches[1], matches[2], nil
}
