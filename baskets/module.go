package baskets

import (
	"context"
	"database/sql"

	"github.com/sanLimbu/eda-go/baskets/basketspb"
	"github.com/sanLimbu/eda-go/baskets/internal/constants"
	"github.com/sanLimbu/eda-go/baskets/internal/domain"
	"github.com/sanLimbu/eda-go/baskets/internal/grpc"
	"github.com/sanLimbu/eda-go/baskets/internal/postgres"
	"github.com/sanLimbu/eda-go/internal/am"
	"github.com/sanLimbu/eda-go/internal/amotel"
	"github.com/sanLimbu/eda-go/internal/amprom"
	"github.com/sanLimbu/eda-go/internal/ddd"
	"github.com/sanLimbu/eda-go/internal/di"
	"github.com/sanLimbu/eda-go/internal/jetstream"
	pg "github.com/sanLimbu/eda-go/internal/postgres"
	"github.com/sanLimbu/eda-go/internal/postgresotel"
	"github.com/sanLimbu/eda-go/internal/registry"
	"github.com/sanLimbu/eda-go/internal/system"
	"github.com/sanLimbu/eda-go/internal/tm"
)

type Module struct{}

func (m *Module) Startup(ctx context.Context, mono system.Service) (err error) {
	return Root(ctx, mono)
}

func Root(ctx context.Context, svc system.Service) (err error) {

	container := di.New()
	//setup Driven adaptors
	container.AddSingleton(constants.RegistryKey, func(c di.Container) (any, error) {
		reg := registry.New()
		if err := domain.Registrations(reg); err != nil {
			return nil, err
		}
		if err := basketspb.Registrations(reg); err != nil {
			return nil, err
		}

		// if err := storespb.Registrations(reg); err != nil {
		// 	return nil, err
		// }
		return reg, nil
	})

	stream := jetstream.NewStream(svc.Config().Nats.Stream, svc.JS(), svc.Logger())
	container.AddSingleton(constants.DomainDispatcherKey, func(c di.Container) (any, error) {
		return ddd.NewEventDispatcher[ddd.Event](), nil
	})
	container.AddScoped(constants.DatabaseTransactionKey, func(c di.Container) (any, error) {
		return svc.DB().Begin()
	})

	sentCounter := amprom.SentMessagesCounter(constants.ServiceName)
	container.AddScoped(constants.MessagePublisherKey, func(c di.Container) (any, error) {
		tx := postgresotel.Trace(c.Get(constants.DatabaseTransactionKey).(*sql.Tx))
		outboxStore := pg.NewOutboxStore(constants.OutboxTableName, tx)
		return am.NewMessagePublisher(
			stream,
			amotel.OtelMessageContextInjector(),
			sentCounter,
			tm.OutboxPublisher(outboxStore),
		), nil
	})

	container.AddSingleton(constants.MessagePublisherKey, func(c di.Container) (any, error) {
		return am.NewMessageSubscriber(
			stream,
			amotel.OtelMessageContextExtractor(),
			amprom.ReceivedMessagesCounter(constants.ServiceName),
		), nil
	})

	container.AddScoped(constants.EventPublisherKey, func(c di.Container) (any, error) {
		return am.NewEventPublisher(
			c.Get(constants.RegistryKey).(registry.Registry),
			c.Get(constants.MessagePublisherKey).(am.MessagePublisher),
		), nil
	})
	// container.AddScoped(constants.InboxStoreKey, func(c di.Container) (any, error) {
	// 	tx := postgresotel.Trace(c.Get(constants.DatabaseTransactionKey).(*sql.Tx))
	// 	return pg.NewInboxStore(constants.InboxTableName, tx), nil
	// })
	// container.AddScoped(constants.BasketsRepoKey, func(c di.Container) (any, error) {
	// 	tx := postgresotel.Trace(c.Get(constants.DatabaseTransactionKey).(*sql.Tx))
	// 	reg := c.Get(constants.RegistryKey).(registry.Registry)
	// 	return es.NewAggregateRepository[*domain.Basket](
	// 		domain.BasketAggregate,
	// 		reg,
	// 		es.AggregateStoreWithMiddleware(
	// 			pg.NewEventStore(constants.EventsTableName, tx, reg),
	// 			pg.NewSnapshotStore(constants.SnapshotsTableName, tx, reg),
	// 		),
	// 	), nil
	// })
	// container.AddScoped(constants.StoresRepoKey, func(c di.Container) (any, error) {
	// 	return postgres.NewStoreCacheRepository(
	// 		constants.StoresCacheTableName,
	// 		postgresotel.Trace(c.Get(constants.DatabaseTransactionKey).(*sql.Tx)),
	// 		grpc.NewStoreRepository(svc.Config().Rpc.Service(constants.StoresServiceName)),
	// 	), nil
	// })
	container.AddScoped(constants.ProductsRepoKey, func(c di.Container) (any, error) {
		return postgres.NewProductCacheRepository(
			constants.ProductsCacheTableName,
			postgresotel.Trace(c.Get(constants.DatabaseTransactionKey).(*sql.Tx)),
			grpc.NewProductRepository(svc.Config().Rpc.Service(constants.StoresServiceName)),
		), nil
	})

	return
}
