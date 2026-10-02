package notifications

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/solargate/grom/internal/users"
	"github.com/solargate/grom/internal/workouts"
)

// Notifier builds events and delivers them to a recipient's devices.
type Notifier struct {
	users    users.Repository
	workouts *workouts.Service
	sender   *Sender
}

func NewNotifier(userStore users.Repository, workoutSvc *workouts.Service, sender *Sender) *Notifier {
	return &Notifier{
		users:    userStore,
		workouts: workoutSvc,
		sender:   sender,
	}
}

// NotifyWorkoutLiked notifies the workout owner about a new like.
func (n *Notifier) NotifyWorkoutLiked(ownerNickname string, workoutID string, actor workouts.WorkoutLikeUser) {
	n.notify(ownerNickname, workoutID, actor, TypeWorkoutLiked, SlotLiked)
}

// NotifyWorkoutCommented notifies the workout owner about a new comment.
func (n *Notifier) NotifyWorkoutCommented(ownerNickname string, workoutID string, actor workouts.WorkoutLikeUser) {
	n.notify(ownerNickname, workoutID, actor, TypeWorkoutCommented, SlotCommented)
}

func (n *Notifier) notify(
	ownerNickname string,
	workoutID string,
	actor workouts.WorkoutLikeUser,
	eventType string,
	slotFn func(owner, workoutID string) string,
) {
	if n == nil || n.sender == nil || n.users == nil || n.workouts == nil {
		return
	}
	ownerNickname = strings.TrimSpace(ownerNickname)
	workoutID = strings.TrimSpace(workoutID)
	if ownerNickname == "" || workoutID == "" {
		return
	}
	if actor.Handle == "" && actor.Nickname == "" && actor.Name == "" {
		return
	}

	// Never notify the owner about their own like/comment (local actors only).
	if actor.IsLocal && strings.EqualFold(strings.TrimSpace(actor.Nickname), ownerNickname) {
		return
	}

	owner, err := n.users.FindByNickname(ownerNickname)
	if err != nil || owner == nil {
		slog.Debug("notifications skip missing owner", "owner", ownerNickname, "err", err)
		return
	}

	workout, err := n.workouts.Get(ownerNickname, workoutID)
	if err != nil || workout == nil {
		slog.Debug("notifications skip missing workout", "owner", ownerNickname, "workout_id", workoutID, "err", err)
		return
	}
	title := strings.TrimSpace(workout.Name)
	if title == "" {
		title = workoutID
	}

	event := Event{
		Type:             eventType,
		ActorDisplayName: ActorDisplayName(actor.Name, actor.Nickname, actor.Handle),
		WorkoutID:        workoutID,
		WorkoutTitle:     title,
		Owner:            ownerNickname,
		Slot:             slotFn(ownerNickname, workoutID),
		EventID:          uuid.NewString(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	n.sender.SendToUser(ctx, owner.ID, event)
}
