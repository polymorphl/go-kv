package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/pubsub"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// Unsubscribe handles the UNSUBSCRIBE command.
// Usage: UNSUBSCRIBE [channel [channel ...]]
// Returns: Array of unsubscribed channels and the number of remaining subscribed channels.
//
// This command unsubscribes the client from the specified channels.
// If no channels are specified, unsubscribes from all channels.
// If the client has no remaining subscriptions, it exits subscribed mode.
//
// Examples:
//
//	UNSUBSCRIBE mychannel1 mychannel2   // Unsubscribe from two channels
//	UNSUBSCRIBE                         // Unsubscribe from all channels
func Unsubscribe(connID string, args []shared.Value) shared.Value {
	// Get current subscriptions
	channels, hasSubscriptions := pubsub.SubscriptionsGet(connID)
	if !hasSubscriptions {
		// Client has no subscriptions, return empty response
		return shared.Value{Typ: "array", Array: []shared.Value{
			{Typ: "bulk", Bulk: "unsubscribe"},
			{Typ: "bulk", Bulk: ""},
			{Typ: "integer", Num: 0},
		}}
	}

	// If no channels specified, unsubscribe from all
	if len(args) == 0 {
		// Unsubscribe from all channels
		pubsub.SubscriptionsDelete(connID)
		pubsub.SubscribedModeDelete(connID)

		// Return response for each previously subscribed channel
		var responses []shared.Value
		for _, channel := range channels {
			responses = append(responses, shared.Value{Typ: "array", Array: []shared.Value{
				{Typ: "bulk", Bulk: "unsubscribe"},
				{Typ: "bulk", Bulk: channel},
				{Typ: "integer", Num: 0},
			}})
		}

		// Return the first response (Redis behavior)
		if len(responses) > 0 {
			return responses[0]
		}

		return shared.Value{Typ: "array", Array: []shared.Value{
			{Typ: "bulk", Bulk: "unsubscribe"},
			{Typ: "bulk", Bulk: ""},
			{Typ: "integer", Num: 0},
		}}
	}

	// Unsubscribe from specified channels
	var unsubscribedChannels []string
	for _, arg := range args {
		channel := arg.Bulk
		// Remove channel from subscriptions if it exists
		channels, _ := pubsub.SubscriptionsGet(connID)
		var newChannels []string
		found := false
		for _, existingChannel := range channels {
			if existingChannel == channel {
				found = true
				unsubscribedChannels = append(unsubscribedChannels, channel)
			} else {
				newChannels = append(newChannels, existingChannel)
			}
		}

		if found {
			if len(newChannels) == 0 {
				// No more subscriptions, remove from subscribed mode
				pubsub.SubscriptionsDelete(connID)
				pubsub.SubscribedModeDelete(connID)
			} else {
				// Update subscriptions with remaining channels
				pubsub.SubscriptionsSetChannels(connID, newChannels)
			}
		}
	}

	// Get remaining subscription count
	remainingChannels, _ := pubsub.SubscriptionsGet(connID)
	remainingCount := len(remainingChannels)

	// Return response for each unsubscribed channel
	var responses []shared.Value
	for _, channel := range unsubscribedChannels {
		responses = append(responses, shared.Value{Typ: "array", Array: []shared.Value{
			{Typ: "bulk", Bulk: "unsubscribe"},
			{Typ: "bulk", Bulk: channel},
			{Typ: "integer", Num: remainingCount},
		}})
	}

	// For single channel unsubscription, return the response directly
	if len(responses) == 1 {
		return responses[0]
	}

	// For multiple channels, return the first response
	if len(responses) > 0 {
		return responses[0]
	}

	// No channels were unsubscribed (they weren't subscribed)
	return shared.Value{Typ: "array", Array: []shared.Value{
		{Typ: "bulk", Bulk: "unsubscribe"},
		{Typ: "bulk", Bulk: ""},
		{Typ: "integer", Num: remainingCount},
	}}
}
