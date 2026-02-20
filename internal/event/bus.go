package event

import (
	"sync"
	"sync/atomic"
)

type Event struct {
	Type    string `json:"type"`
	BoardID string `json:"board_id"`
	Payload any    `json:"payload"`
	SeqID   uint64 `json:"seq_id"`
}

type Subscriber struct {
	Ch      chan Event
	boardID string
}

type Bus struct {
	mu   sync.RWMutex
	subs map[string][]*Subscriber
	seq  atomic.Uint64
}

func NewBus() *Bus {
	return &Bus{
		subs: make(map[string][]*Subscriber),
	}
}

func (b *Bus) Subscribe(boardID string) *Subscriber {
	sub := &Subscriber{
		Ch:      make(chan Event, 64),
		boardID: boardID,
	}
	b.mu.Lock()
	b.subs[boardID] = append(b.subs[boardID], sub)
	b.mu.Unlock()
	return sub
}

func (b *Bus) Unsubscribe(sub *Subscriber) {
	b.mu.Lock()
	defer b.mu.Unlock()
	subs := b.subs[sub.boardID]
	for i, s := range subs {
		if s == sub {
			b.subs[sub.boardID] = append(subs[:i], subs[i+1:]...)
			close(sub.Ch)
			return
		}
	}
}

func (b *Bus) Publish(evt Event) {
	evt.SeqID = b.seq.Add(1)
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, sub := range b.subs[evt.BoardID] {
		select {
		case sub.Ch <- evt:
		default:
		}
	}
}
