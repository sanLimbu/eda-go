package handlers

import (
	"context"
	"database/sql"

	"github.com/sanLimbu/eda-go/customers/internal/constants"
	"github.com/sanLimbu/eda-go/internal/am"
	"github.com/sanLimbu/eda-go/internal/di"
)

func RegisterCommandHandlersTx(container di.Container) error {
	rawMsgHandler := am.MessageHandlerFunc(func(ctx context.Context, msg am.IncomingMessage) (err error) {
		ctx = container.Scoped(ctx)
		defer func(tx *sql.Tx) {
			if p := recover(); p != nil {
				_ = tx.Rollback()
				panic(p)
			} else if err != nil {
				_ = tx.Rollback()
			} else {
				err = tx.Commit()
			}

		}(di.Get(ctx, constants.DatabaseTransactionKey).(*sql.Tx))
		return di.Get(ctx, constants.CommandHandlersKey).(am.MessageHandler).HandleMessage(ctx, msg)

	})

	subscriber := container.Get(constants.MessageSubscriberKey).(am.MessageSubscriber)
	return RegisterCommandHandlers(subscriber, rawMsgHandler)
}
