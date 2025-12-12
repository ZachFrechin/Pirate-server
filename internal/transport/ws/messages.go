package ws

// ClientMessage is any message coming from Unity/client.
type ClientMessage struct {
	Type      string `json:"type"`
	Name      string `json:"name,omitempty"`
	Code      string `json:"code,omitempty"`
	HandIndex int    `json:"handIndex,omitempty"`
	TargetID  string `json:"targetId,omitempty"`
}

// ServerMessage is any message sent from server to client.
type ServerMessage struct {
	Type     string      `json:"type"`
	Message  string      `json:"message,omitempty"`
	Code     string      `json:"code,omitempty"`
	PlayerID string      `json:"playerId,omitempty"`
	State    interface{} `json:"state,omitempty"`
}
