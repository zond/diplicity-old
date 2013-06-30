package common

const (
	SubscribeType   = "subscribe"
	UnsubscribeType = "unsubscribe"
)

type SubscribeMessage struct {
	URI string
}

type ObjectMessage struct {
	Data interface{}
	URL  string
}

type JsonMessage struct {
	Type      string            `json:",omitempty"`
	Subscribe *SubscribeMessage `json:",omitempty"`
	Object    *ObjectMessage    `json:",omitempty"`
}
