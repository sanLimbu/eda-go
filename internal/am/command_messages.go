package am

import (
	"context"
	"time"

	"github.com/sanLimbu/eda-go/internal/ddd"
	"github.com/sanLimbu/eda-go/internal/registry"
)

const (
	CommandHdrPrefix       = "COMMAND_"
	CommandNameHdr         = CommandHdrPrefix + "NAME"
	CommandReplyChannelHdr = CommandHdrPrefix + "REPLY_CHANNEL"
)

type (
	CommandMessage interface {
		MessageBase
		ddd.Command
	}

	IncomingCommandMessage interface {
		IncomingMessageBase
		ddd.Command
	}

	CommandPublisher interface {
		Publish(ctx context.Context, topicName string, cmd ddd.Command) error
	}

	commandPublisher struct {
		reg       registry.Registry
		publisher MessagePublisher
	}

	commandMessage struct {
		id         string
		name       string
		payload    ddd.CommandPayload
		occurredAt time.Time
		msg        IncomingMessageBase
	}

	commandMsgHandler struct {
		reg       registry.Registry
		publisher ReplyPublisher
		handler   ddd.CommandHandler[ddd.Command]
	}
)
