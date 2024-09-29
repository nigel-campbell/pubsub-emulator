package pstest

import (
	"gotest.tools/v3/assert"
	"testing"
)

func TestRowKey(t *testing.T) {
	topicKey := topicRowKey("foo")
	topicStr, err := parseTopicRowKey(topicKey)
	assert.Assert(t, err == nil, "failed to parse topic row key")
	assert.Assert(t, topicStr == "foo", "parsed incorrect topic row key")

	subKey := subscriptionRowKey("baz", "bat")
	topicStr, subStr, err := parseSubscriptionRowKey(subKey)
	assert.Assert(t, err == nil, "failed to parse subscription row key")
	assert.Assert(t, topicStr == "baz", "parsed incorrect topic row key")
	assert.Assert(t, subStr == "bat", "parsed incorrect subscription row key")
}
