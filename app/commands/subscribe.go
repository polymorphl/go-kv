package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/pubsub"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

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

	// Register subscriptions for all channels
	for _, arg := range args {
		channel := arg.Bulk
		pubsub.SubscriptionsSet(connID, channel)
	}

	// Set client in subscribed mode
	pubsub.SubscribedModeSet(connID)

	// Get current subscription count
	channels, _ := pubsub.SubscriptionsGet(connID)
	subscriptionCount := len(channels)

	// Return response for each subscribed channel
	var responses []shared.Value
	for _, channel := range args {
		responses = append(responses, shared.Value{Typ: "array", Array: []shared.Value{
			{Typ: "bulk", Bulk: "subscribe"},
			{Typ: "bulk", Bulk: channel.Bulk},
			{Typ: "integer", Num: subscriptionCount},
		}})
	}

	// For single channel subscription, return the response directly
	if len(responses) == 1 {
		return responses[0]
	}

	// For multiple channels, return the first response
	return responses[0]
}
