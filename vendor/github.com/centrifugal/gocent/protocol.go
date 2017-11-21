package gocent

import (
	"encoding/json"
)

// Command represents API command to send.
type Command struct {
	UID    string                 `json:"uid"`
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
}

// Response is a response of server on command sent.
type Response struct {
	Method string          `json:"method"`
	Error  string          `json:"error"`
	Body   json.RawMessage `json:"body"`
}

// Result is a slice of responses.
type Result []Response

// ClientInfo represents information about one client connection to Centrifugo.
// This struct used in messages published by clients, join/leave events, presence data.
type ClientInfo struct {
	User        string           `json:"user"`
	Client      string           `json:"client"`
	DefaultInfo *json.RawMessage `json:"default_info,omitempty"`
	ChannelInfo *json.RawMessage `json:"channel_info,omitempty"`
}

// Message represents message published into channel.
type Message struct {
	UID       string           `json:"uid"`
	Timestamp string           `json:"timestamp"`
	Info      *ClientInfo      `json:"info,omitempty"`
	Channel   string           `json:"channel"`
	Data      *json.RawMessage `json:"data"`
	Client    string           `json:"client,omitempty"`
}

// NodeInfo contains information and statistics about Centrifugo node.
type NodeInfo struct {
	// UID is a unique id of running node.
	UID string `json:"uid"`
	// Name is a name of node (config defined or generated automatically).
	Name string `json:"name"`
	// Started is node start timestamp.
	Started int64 `json:"started_at"`
	// Metrics contains Centrifugo metrics.
	Metrics map[string]int64 `json:"metrics"`
}

// Stats contains state and metrics information from all running Centrifugo nodes.
type Stats struct {
	Nodes           []NodeInfo `json:"nodes"`
	MetricsInterval int64      `json:"metrics_interval"`
}

// presenceBody represents body of response in case of successful presence command.
type presenceBody struct {
	Channel string                `json:"channel"`
	Data    map[string]ClientInfo `json:"data"`
}

// historyBody represents body of response in case of successful history command.
type historyBody struct {
	Channel string    `json:"channel"`
	Data    []Message `json:"data"`
}

// channelsBody represents body of response in case of successful channels command.
type channelsBody struct {
	Data []string `json:"data"`
}

// statsBody represents body of response in case of successful stats command.
type statsBody struct {
	Data Stats `json:"data"`
}
