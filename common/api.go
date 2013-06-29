package common

const (
	SubscribeType   = "subscribe"
	UnsubscribeType = "unsubscribe"
)

type SubscribeMessage struct {
	URI string
}

type JsonMessage struct {
	Type      string
	Subscribe *SubscribeMessage
	Object    interface{}
}
