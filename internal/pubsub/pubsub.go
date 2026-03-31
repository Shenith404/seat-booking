package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

// EventType represents the type of event
type EventType string

const (
	EventSeatHeld     EventType = "SEAT_HELD"
	EventSeatReleased EventType = "SEAT_RELEASED"
	EventSeatBooked   EventType = "SEAT_BOOKED"
)

// Event represents a pub/sub event
type Event struct {
	Type    EventType `json:"type"`
	ShowID  string    `json:"show_id"`
	SeatID  string    `json:"seat_id,omitempty"`
	SeatIDs []string  `json:"seat_ids,omitempty"`
	Data    any       `json:"data,omitempty"`
}

// PubSub handles Redis pub/sub operations
type PubSub struct {
	client *redis.Client
}

// NewPubSub creates a new PubSub instance
func NewPubSub(client *redis.Client) *PubSub {
	return &PubSub{client: client}
}

// channelName returns the channel name for a show
func channelName(showID string) string {
	return fmt.Sprintf("show_updates:%s", showID)
}

// Publish publishes an event to a show's channel
func (ps *PubSub) Publish(ctx context.Context, event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	channel := channelName(event.ShowID)
	if err := ps.client.Publish(ctx, channel, data).Err(); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	log.Printf("Published event %s to channel %s", event.Type, channel)
	return nil
}

// Subscribe subscribes to a show's channel and returns messages
func (ps *PubSub) Subscribe(ctx context.Context, showID string) (<-chan Event, func()) {
	channel := channelName(showID)
	sub := ps.client.Subscribe(ctx, channel)

	eventChan := make(chan Event, 100)

	go func() {
		defer close(eventChan)
		ch := sub.Channel()

		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}

				var event Event
				if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
					log.Printf("Failed to unmarshal event: %v", err)
					continue
				}

				select {
				case eventChan <- event:
				default:
					log.Printf("Event channel full, dropping event")
				}
			}
		}
	}()

	cleanup := func() {
		sub.Close()
	}

	return eventChan, cleanup
}

// PublishSeatHeld publishes a seat held event
func (ps *PubSub) PublishSeatHeld(ctx context.Context, showID, seatID string) error {
	return ps.Publish(ctx, Event{
		Type:   EventSeatHeld,
		ShowID: showID,
		SeatID: seatID,
	})
}

// PublishSeatReleased publishes a seat released event
func (ps *PubSub) PublishSeatReleased(ctx context.Context, showID, seatID string) error {
	return ps.Publish(ctx, Event{
		Type:   EventSeatReleased,
		ShowID: showID,
		SeatID: seatID,
	})
}

// PublishSeatsBooked publishes a seats booked event
func (ps *PubSub) PublishSeatsBooked(ctx context.Context, showID string, seatIDs []string) error {
	return ps.Publish(ctx, Event{
		Type:    EventSeatBooked,
		ShowID:  showID,
		SeatIDs: seatIDs,
	})
}
