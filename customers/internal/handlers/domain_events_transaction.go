package handlers

import (
	"context"

	"github.com/sanLimbu/eda-go/customers/internal/constants"
	"github.com/sanLimbu/eda-go/internal/ddd"
	"github.com/sanLimbu/eda-go/internal/di"
)

func RegisterDomainEventHandlersTx(container di.Container) {
	handlers := ddd.EventHandlerFunc[ddd.AggregateEvent](func(ctx context.Context, event ddd.AggregateEvent) error {

		domainHandlers := di.Get(ctx, constants.DomainEventHandlersKey).(ddd.EventHandler[ddd.AggregateEvent])
		return domainHandlers.HandleEvent(ctx, event)
	})

	subscriber := container.Get(constants.DomainDispatcherKey).(*ddd.EventDispatcher[ddd.AggregateEvent])
	RegisterDomainEventHandlers(subscriber, handlers)
}
