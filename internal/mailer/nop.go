package mailer

import "context"

// Nop is a disabled mailer.
type Nop struct{}

func (Nop) Send(context.Context, Message) error {
	return ErrDisabled
}
