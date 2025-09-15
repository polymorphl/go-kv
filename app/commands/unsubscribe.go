package commands

import (
	"sync"

	"github.com/codecrafters-io/redis-starter-go/app/pubsub"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// Object pool for response slices to reduce allocations
var (
	unsubscribeResponsePool = sync.Pool{
		New: func() interface{} {
			return make([]shared.Value, 0, 8)
		},
	}
	unsubscribeChannelPool = sync.Pool{
		New: func() interface{} {
			return make([]string, 0, 16)
		},
	}
)

// Helper functions for pool management
func getUnsubscribeResponse() []shared.Value {
	return unsubscribeResponsePool.Get().([]shared.Value)
}

func putUnsubscribeResponse(s []shared.Value) {
	s = s[:0] // Reset length but keep capacity
	unsubscribeResponsePool.Put(s)
}

func getUnsubscribeChannelSlice() []string {
	return unsubscribeChannelPool.Get().([]string)
}

func putUnsubscribeChannelSlice(s []string) {
	s = s[:0] // Reset length but keep capacity
	unsubscribeChannelPool.Put(s)
}

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
	// Get current subscriptions once
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

		// Use object pool for responses
		responses := getUnsubscribeResponse()
		defer putUnsubscribeResponse(responses)

		// Pre-allocate with exact capacity
		responses = responses[:0]
		if cap(responses) < len(channels) {
			responses = make([]shared.Value, 0, len(channels))
		}

		// Create responses efficiently
		for _, channel := range channels {
			responseArray := []shared.Value{
				{Typ: "bulk", Bulk: "unsubscribe"},
				{Typ: "bulk", Bulk: channel},
				{Typ: "integer", Num: 0},
			}
			responses = append(responses, shared.Value{Typ: "array", Array: responseArray})
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

	// Use object pools for channel slices
	unsubscribedChannels := getUnsubscribeChannelSlice()
	defer putUnsubscribeChannelSlice(unsubscribedChannels)

	// Pre-allocate with estimated capacity
	unsubscribedChannels = unsubscribedChannels[:0]
	if cap(unsubscribedChannels) < len(args) {
		unsubscribedChannels = make([]string, 0, len(args))
	}

	// Create a map of channels to unsubscribe for efficient lookup
	channelsToUnsubscribe := make(map[string]bool, len(args))
	for _, arg := range args {
		channelsToUnsubscribe[arg.Bulk] = true
	}

	// Build new channels list efficiently
	newChannels := make([]string, 0, len(channels))
	for _, existingChannel := range channels {
		if channelsToUnsubscribe[existingChannel] {
			unsubscribedChannels = append(unsubscribedChannels, existingChannel)
		} else {
			newChannels = append(newChannels, existingChannel)
		}
	}

	// Update subscriptions
	if len(newChannels) == 0 {
		// No more subscriptions, remove from subscribed mode
		pubsub.SubscriptionsDelete(connID)
		pubsub.SubscribedModeDelete(connID)
	} else {
		// Update subscriptions with remaining channels
		pubsub.SubscriptionsSetChannels(connID, newChannels)
	}

	remainingCount := len(newChannels)

	// Use object pool for responses
	responses := getUnsubscribeResponse()
	defer putUnsubscribeResponse(responses)

	// Pre-allocate with exact capacity
	responses = responses[:0]
	if cap(responses) < len(unsubscribedChannels) {
		responses = make([]shared.Value, 0, len(unsubscribedChannels))
	}

	// Create responses efficiently
	for _, channel := range unsubscribedChannels {
		responseArray := []shared.Value{
			{Typ: "bulk", Bulk: "unsubscribe"},
			{Typ: "bulk", Bulk: channel},
			{Typ: "integer", Num: remainingCount},
		}
		responses = append(responses, shared.Value{Typ: "array", Array: responseArray})
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
