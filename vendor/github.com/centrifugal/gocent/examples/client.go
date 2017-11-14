package main

import (
	"fmt"
	"time"

	"github.com/centrifugal/gocent"
)

func main() {

	ch := "$public:chat"

	c := gocent.NewClient("http://localhost:8000", "secret", 5*time.Second)

	// How to publish.
	ok, err := c.Publish(ch, []byte(`{"input": "test"}`))
	if err != nil {
		println(err.Error())
		return
	}
	fmt.Printf("Publish into channel %s successful: %v\n", ch, ok)

	// How to get presence.
	presence, _ := c.Presence(ch)
	fmt.Printf("Presense for channel %s: %v\n", ch, presence)

	// How to get history.
	history, _ := c.History(ch)
	fmt.Printf("History for channel %s, %d messages: %v\n", ch, len(history), history)

	// How to get channels.
	channels, _ := c.Channels()
	fmt.Printf("Channels: %v\n", channels)

	// How to export stats.
	stats, _ := c.Stats()
	fmt.Printf("Stats: %v\n", stats)

	// How to send 3 commands in one request.
	_ = c.AddPublish(ch, []byte(`{"input": "test1"}`))
	_ = c.AddPublish(ch, []byte(`{"input": "test2"}`))
	_ = c.AddPublish(ch, []byte(`{"input": "test3"}`))
	result, err := c.Send()
	fmt.Println("Sent", len(result), "publish commands in one request")

	// How to broadcast the same data into 3 different channels in one request.
	chs := []string{"$public:chat_1", "$public:chat_2", "$public:chat_3"}
	ok, err = c.Broadcast(chs, []byte(`{"input": "test"}`))
	if err != nil {
		println(err.Error())
		return
	}
	println("Broadcast successful:", ok)
}
