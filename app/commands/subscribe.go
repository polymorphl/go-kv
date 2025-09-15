package commands

import (
	"sync"

	"github.com/codecrafters-io/redis-starter-go/app/pubsub"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// Object pool for response slices to reduce allocations
var (
	subscribeResponsePool = sync.Pool{
		New: func() interface{} {
			return make([]shared.Value, 0, 8)
		},
	}
)

// Helper functions for pool management
func getSubscribeResponse() []shared.Value {
	return subscribeResponsePool.Get().([]shared.Value)
}

func putSubscribeResponse(s []shared.Value) {
	s = s[:0] // Reset length but keep capacity
	subscribeResponsePool.Put(s)
}

// Subscribe handles the SUBSCRIBE command.
// Usage: SUBSCRIBE channel [channel ...]
// Returns: Array of subscribed channels and the number of subscribed channels.
//
// This command registers the client to listen for messages published to the specified channels.
// The client will receive messages published to any of the subscribed channels.
//
// Examples:
//
//	SUBSCRIBE mychannel1 mychannel2   // Subscribe to two channels
func Subscribe(connID string, args []shared.Value) shared.Value {
	if len(args) == 0 {
		return createErrorResponse("ERR wrong number of arguments for 'subscribe' command")
	}

	// Register subscriptions for all channels efficiently
	newChannels := make([]string, 0, len(args))
	for _, arg := range args {
		channel := arg.Bulk
		pubsub.SubscriptionsSet(connID, channel)
		newChannels = append(newChannels, channel)
	}

	// Set client in subscribed mode
	pubsub.SubscribedModeSet(connID)

	// Get current subscription count
	channels, _ := pubsub.SubscriptionsGet(connID)
	subscriptionCount := len(channels)

	// Use object pool for response slice
	responses := getSubscribeResponse()
	defer putSubscribeResponse(responses)

	// Pre-allocate with exact capacity
	responses = responses[:0]
	if cap(responses) < len(newChannels) {
		responses = make([]shared.Value, 0, len(newChannels))
	}

	// Create responses efficiently
	for _, channel := range newChannels {
		responseArray := []shared.Value{
			{Typ: "bulk", Bulk: "subscribe"},
			{Typ: "bulk", Bulk: channel},
			{Typ: "integer", Num: subscriptionCount},
		}
		responses = append(responses, shared.Value{Typ: "array", Array: responseArray})
	}

	// For single channel subscription, return the response directly
	if len(responses) == 1 {
		return responses[0]
	}

	// For multiple channels, return the first response
	return responses[0]
}
