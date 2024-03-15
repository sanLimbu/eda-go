package system

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/nats.go"
	"github.com/pressly/goose/v3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/sanLimbu/eda-go/internal/config"
	"github.com/sanLimbu/eda-go/internal/logger"
	"github.com/stackus/errors"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type System struct {
	cfg config.AppConfig
	db  *sql.DB
	nc  *nats.Conn
	js  nats.JetStreamContext
	mux *chi.Mux
	rpc *grpc.Server
	//waiter waiter.Waiter
	logger zerolog.Logger
	tp     *sdktrace.TracerProvider
}

func NewSystem(cfg config.AppConfig) (*System, error) {
	s := &System{cfg: cfg}
	//s.initWaiter()
	if err := s.initDB(); err != nil {
		return nil, err
	}
	// if err := s.initJS(); err != nil {
	// 	return nil, err
	// }

	// if err := s.initOpenTelemetry(); err != nil {
	// 	return nil, err
	// }

	s.initMux()
	s.initRpc()
	s.initLogger()

	return s, nil
}

func (s *System) Config() config.AppConfig {
	return s.cfg
}

func (s *System) initDB() (err error) {
	s.db, err = sql.Open("pgx", s.cfg.PG.Conn)
	return err
}

func (s *System) MigrateDB(fs fs.FS) error {
	goose.SetBaseFS(fs)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	if err := goose.Up(s.db, "."); err != nil {
		return err
	}
	return nil

}

func (s *System) DB() *sql.DB {
	return s.db
}

func (s *System) initJS() (err error) {
	s.nc, err = nats.Connect(s.cfg.Nats.URL)
	if err != nil {
		return err
	}

	s.js, err = s.nc.JetStream()
	if err != nil {
		return err
	}

	_, err = s.js.AddStream(&nats.StreamConfig{
		Name:     s.cfg.Nats.Stream,
		Subjects: []string{fmt.Sprintf("%s.>", s.cfg.Nats.Stream)},
	})
	return err

}

func (s *System) JS() nats.JetStreamContext {
	return s.js
}

func (s *System) initLogger() {
	s.logger = logger.New(logger.LogConfig{
		Environment: s.cfg.Environment,
		LogLevel:    logger.Level(s.cfg.LogLevel),
	})
}

func (s *System) Logger() zerolog.Logger {
	return s.logger
}

func (s *System) initMux() {
	s.mux = chi.NewMux()
	s.mux.Use(middleware.Heartbeat("/liveness"))
	s.mux.Method("GET", "/metrics", promhttp.Handler())
}

func (s *System) Mux() *chi.Mux {
	return s.mux
}

func (s *System) initRpc() {
	s.rpc = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			otelgrpc.UnaryServerInterceptor(),
			serverErrorUnaryInterceptor(),
		),

		//If there are streaming endpoints also add
		// grpc.StreamInterceptor(
		// 	otelgrpc.StreamServerInterceptor(),
		// ),
	)
	reflection.Register(s.rpc)
}

func (s *System) RPC() *grpc.Server {
	return s.rpc
}

// func (s *System) initWaiter() {
// 	s.waiter = waiter.New(waiter.CatchSignals())
// }

// func (s *System) Waiter() waiter.Waiter {
// 	return s.waiter
// }

func (s *System) WaitForWeb(ctx context.Context) error {
	webServer := &http.Server{
		Addr:    s.cfg.Web.Address(),
		Handler: s.mux,
	}

	group, gCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		fmt.Printf("web server started; listening at http://localhost%s\n", s.cfg.Web.Port)
		defer fmt.Println("web server shutdown")
		if err := webServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	})

	group.Go(func() error {
		<-gCtx.Done()
		fmt.Println("web server to be shut down")
		ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
		defer cancel()
		if err := webServer.Shutdown(ctx); err != nil {
			return err
		}
		return nil
	})
	return group.Wait()

}

func serverErrorUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		resp, err = handler(ctx, req)
		return resp, errors.SendGRPCError(err)
	}
}