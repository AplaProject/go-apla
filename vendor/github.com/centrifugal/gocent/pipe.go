package gocent

import (
	"encoding/json"
	"sync"
)

// Pipe allows to send several commands in one HTTP request.
type Pipe struct {
	mu   sync.RWMutex
	cmds []Command
}

// Reset allows to clear client command buffer.
func (p *Pipe) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cmds = nil
}

func (p *Pipe) add(cmd Command) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cmds = append(p.cmds, cmd)
	return nil
}

// AddPublish adds publish command to client command buffer but not actually
// sends request to server until Pipe will be explicitly sent.
func (p *Pipe) AddPublish(channel string, data []byte) error {
	var raw json.RawMessage
	raw = json.RawMessage(data)
	cmd := Command{
		Method: "publish",
		Params: map[string]interface{}{
			"channel": channel,
			"data":    &raw,
		},
	}
	return p.add(cmd)
}

// AddBroadcast adds broadcast command to client command buffer but not actually
// sends request to server until Pipe will be explicitly sent.
func (p *Pipe) AddBroadcast(channels []string, data []byte) error {
	var raw json.RawMessage
	raw = json.RawMessage(data)
	cmd := Command{
		Method: "broadcast",
		Params: map[string]interface{}{
			"channels": channels,
			"data":     &raw,
		},
	}
	return p.add(cmd)
}

// AddUnsubscribe adds unsubscribe command to client command buffer but not actually
// sends request to server until Pipe will be explicitly sent.
func (p *Pipe) AddUnsubscribe(channel string, user string) error {
	cmd := Command{
		Method: "unsubscribe",
		Params: map[string]interface{}{
			"channel": channel,
			"user":    user,
		},
	}
	return p.add(cmd)
}

// AddDisconnect adds disconnect command to client command buffer but not actually
// sends request to server until Pipe will be explicitly sent.
func (p *Pipe) AddDisconnect(user string) error {
	cmd := Command{
		Method: "disconnect",
		Params: map[string]interface{}{
			"user": user,
		},
	}
	return p.add(cmd)
}

// AddPresence adds presence command to client command buffer but not actually
// sends request to server until Pipe will be explicitly sent.
func (p *Pipe) AddPresence(channel string) error {
	cmd := Command{
		Method: "presence",
		Params: map[string]interface{}{
			"channel": channel,
		},
	}
	return p.add(cmd)
}

// AddPresenceStats adds presence stats command to client command buffer but not actually
// sends request to server until Pipe will be explicitly sent.
func (p *Pipe) AddPresenceStats(channel string) error {
	cmd := Command{
		Method: "presence_stats",
		Params: map[string]interface{}{
			"channel": channel,
		},
	}
	return p.add(cmd)
}

// AddHistory adds history command to client command buffer but not actually
// sends request to server until Pipe will be explicitly sent.
func (p *Pipe) AddHistory(channel string) error {
	cmd := Command{
		Method: "history",
		Params: map[string]interface{}{
			"channel": channel,
		},
	}
	return p.add(cmd)
}

// AddChannels adds channels command to client command buffer but not actually
// sends request to server until Pipe will be explicitly sent.
func (p *Pipe) AddChannels() error {
	cmd := Command{
		Method: "channels",
		Params: map[string]interface{}{},
	}
	return p.add(cmd)
}

// AddInfo adds info command to client command buffer but not actually
// sends request to server until Pipe will be explicitly sent.
func (p *Pipe) AddInfo() error {
	cmd := Command{
		Method: "info",
		Params: map[string]interface{}{},
	}
	return p.add(cmd)
}
