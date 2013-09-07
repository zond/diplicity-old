package common

const (
	SubscribeType   = "subscribe"
	UnsubscribeType = "unsubscribe"
	CreateType      = "create"
)

type SubscribeMessage struct {
	URI string
}

type JsonMessage struct {
	Type      string                 `json:",omitempty"`
	Subscribe *SubscribeMessage      `json:",omitempty"`
	Object    map[string]interface{} `json:",omitempty"`
}
