package system

import (
	"database/sql"

	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/sanLimbu/eda-go/internal/config"
	"google.golang.org/grpc"
)

type Service interface {
	Config() config.AppConfig
	DB() *sql.DB
	JS() nats.JetStreamContext
	Mux() *chi.Mux
	RPC() *grpc.Server
	//	Waiter() waiter.Waiter
	Logger() zerolog.Logger
}
