package event_test

import (
	"testing"
	"time"

	"github.com/aellingwood/cielo/internal/event"
)

func TestBus_SubscribeAndPublish(t *testing.T) {
	bus := event.NewBus()
	sub := bus.Subscribe("board-1")

	bus.Publish(event.Event{Type: "card.created", BoardID: "board-1", Payload: "hello"})

	select {
	case evt := <-sub.Ch:
		if evt.Type != "card.created" || evt.Payload != "hello" {
			t.Errorf("unexpected event: %+v", evt)
		}
		if evt.SeqID == 0 {
			t.Error("expected non-zero SeqID")
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
	}
	bus.Unsubscribe(sub)
}

func TestBus_BoardScoping(t *testing.T) {
	bus := event.NewBus()
	sub1 := bus.Subscribe("board-1")
	sub2 := bus.Subscribe("board-2")

	bus.Publish(event.Event{Type: "card.created", BoardID: "board-1", Payload: "msg1"})

	select {
	case <-sub1.Ch:
	case <-time.After(time.Second):
		t.Fatal("sub1 should receive event")
	}

	select {
	case <-sub2.Ch:
		t.Fatal("sub2 should not receive board-1 event")
	case <-time.After(50 * time.Millisecond):
	}

	bus.Unsubscribe(sub1)
	bus.Unsubscribe(sub2)
}

func TestBus_MultipleSubscribers(t *testing.T) {
	bus := event.NewBus()
	sub1 := bus.Subscribe("board-1")
	sub2 := bus.Subscribe("board-1")

	bus.Publish(event.Event{Type: "test", BoardID: "board-1", Payload: "data"})

	for _, sub := range []*event.Subscriber{sub1, sub2} {
		select {
		case evt := <-sub.Ch:
			if evt.Payload != "data" {
				t.Errorf("unexpected payload")
			}
		case <-time.After(time.Second):
			t.Fatal("timeout")
		}
	}
	bus.Unsubscribe(sub1)
	bus.Unsubscribe(sub2)
}
