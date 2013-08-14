package common

const (
	SubscribeType   = "subscribe"
	UnsubscribeType = "unsubscribe"
)

type SubscribeMessage struct {
	URI string
}

type JsonMessage struct {
	Type      string            `json:",omitempty"`
	Subscribe *SubscribeMessage `json:",omitempty"`
}
