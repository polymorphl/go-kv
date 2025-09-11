package commands

import (
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

// Publish handles the PUBLISH command.
// Usage: PUBLISH channel message
// Returns: The number of clients that received the message.
//
// This command publishes a message to the specified channel.
// The message is delivered to all clients that are subscribed to the channel.
//
// Examples:
//
//	PUBLISH mychannel "Hello, Redis!"   // Publish a message to the mychannel
func Publish(connID string, args []shared.Value) shared.Value {
	if len(args) < 2 {
		return createErrorResponse("ERR wrong number of arguments for 'publish' command")
	}

	channel := args[0].Bulk
	// message := args[1].Bulk // Message is not used in this implementation

	// Count the number of clients subscribed to this channel
	subscriberCount := shared.SubscriptionsCountForChannel(channel)

	return shared.Value{Typ: "integer", Num: subscriberCount}
}
