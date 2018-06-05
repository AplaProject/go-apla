package parser

import "sync"

type parserCache struct {
	mutex sync.RWMutex
	cache map[string]*Parser
}

func (pc *parserCache) Get(hash string) (p *Parser, ok bool) {
	pc.mutex.RLock()
	defer pc.mutex.RUnlock()

	p, ok = pc.cache[hash]
	return
}

func (pc *parserCache) Set(p *Parser) {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()

	pc.cache[string(p.TxHash)] = p
}

func (pc *parserCache) Clean() {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()

	pc.cache = make(map[string]*Parser)
}
