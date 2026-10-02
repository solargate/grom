package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"
)

// Sender delivers encrypted Web Push messages.
type Sender struct {
	keys       *VAPIDStore
	subscriber string
	repo       Repository
	client     *http.Client
	ttl        int
}

func NewSender(keys *VAPIDStore, repo Repository, subscriber string) *Sender {
	sub := strings.TrimSpace(subscriber)
	if sub == "" {
		sub = "mailto:noreply@localhost"
	}
	return &Sender{
		keys:       keys,
		subscriber: sub,
		repo:       repo,
		client:     &http.Client{Timeout: 15 * time.Second},
		ttl:        86400,
	}
}

// SendToUser pushes event JSON to every subscription for userID.
func (s *Sender) SendToUser(ctx context.Context, userID string, event Event) {
	if s == nil || s.keys == nil || s.repo == nil || userID == "" {
		return
	}
	keys := s.keys.Keys()
	if keys == nil {
		return
	}
	subs, err := s.repo.ListByUser(userID)
	if err != nil {
		slog.Warn("notifications list subscriptions failed", "user_id", userID, "err", err)
		return
	}
	if len(subs) == 0 {
		return
	}
	payload, err := json.Marshal(event)
	if err != nil {
		slog.Warn("notifications marshal event failed", "err", err)
		return
	}
	for _, sub := range subs {
		if err := s.sendOne(ctx, keys, sub, payload, event.Slot); err != nil {
			slog.Warn("notifications push failed",
				"user_id", userID,
				"installation_id", sub.InstallationID,
				"err", err,
			)
		}
	}
}

func (s *Sender) sendOne(ctx context.Context, keys *VAPIDKeys, sub PushSubscription, payload []byte, topic string) error {
	subscription := &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys: webpush.Keys{
			P256dh: sub.P256dh,
			Auth:   sub.Auth,
		},
	}
	opts := &webpush.Options{
		HTTPClient:      s.client,
		Subscriber:      s.subscriber,
		VAPIDPublicKey:  keys.PublicKey,
		VAPIDPrivateKey: keys.PrivateKey,
		TTL:             s.ttl,
		Topic:           topic,
		Urgency:         webpush.UrgencyNormal,
	}
	resp, err := webpush.SendNotificationWithContext(ctx, payload, subscription, opts)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted:
		return nil
	case http.StatusNotFound, http.StatusGone:
		if delErr := s.repo.DeleteByEndpoint(sub.Endpoint); delErr != nil {
			slog.Warn("notifications delete dead endpoint failed", "endpoint", sub.Endpoint, "err", delErr)
		}
		return fmt.Errorf("push endpoint gone (%d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	default:
		return fmt.Errorf("push status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
}
