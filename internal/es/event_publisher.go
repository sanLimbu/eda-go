package es

import (
	"context"

	"github.com/sanLimbu/eda-go/internal/ddd"
)

type EventPublisher struct {
	AggregateStore
	publisher ddd.EventPublisher[ddd.AggregateEvent]
}

var _ AggregateStore = (*EventPublisher)(nil)

func NewEventPublisher(publisher ddd.EventPublisher[ddd.AggregateEvent]) AggregateStoreMiddleware {

	eventPublisher := EventPublisher{
		publisher: publisher,
	}

	return func(store AggregateStore) AggregateStore {
		eventPublisher.AggregateStore = store
		return eventPublisher
	}
}
func (p EventPublisher) Save(ctx context.Context, aggaggregate EventSourcedAggregate) error {

	if err := p.AggregateStore.Save(ctx, aggaggregate); err != nil {
		return err
	}
	return p.publisher.Publish(ctx, aggaggregate.Events()...)
}
