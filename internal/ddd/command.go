package ddd

import (
	"context"
	"time"
)

type (
	CommandHandler[T Command] interface {
		HandleCommand(ctx context.Context, cmd T) (Reply, error)
	}

	CommandHandlerFunc[T Command] func(ctx context.Context, cmd T) (Reply, error)

	CommandOption interface {
		configureCommand(*command)
	}

	CommandPayload any

	Command interface {
		IDer
		CommandName() string
		Payload() CommandPayload
		Metadata() Metadata
		OccurredAt() time.Time
	}

	command struct {
		Entity
		payload    CommandPayload
		metadata   Metadata
		occurredAt time.Time
	}
)
