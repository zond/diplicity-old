package common

const (
	SubscribeType   = "Subscribe"
	UnsubscribeType = "Unsubscribe"
	UpdateType      = "Update"
	CreateType      = "Create"
)

type SubscribeMessage struct {
	URI string
}

type CreateMessage struct {
	URI    string
	Object map[string]interface{}
}

type UpdateMessage struct {
	URI    string
	Object map[string]interface{}
}

type JsonMessage struct {
	Type      string            `json:",omitempty"`
	Subscribe *SubscribeMessage `json:",omitempty"`
	Create    *CreateMessage    `json:",omitempty"`
	Update    *UpdateMessage    `json:",omitempty"`
}
