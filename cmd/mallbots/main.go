package mallbots

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/sanLimbu/eda-go/baskets"
	"github.com/sanLimbu/eda-go/baskets/migrations"
	"github.com/sanLimbu/eda-go/customers"
	"github.com/sanLimbu/eda-go/internal/config"
	"github.com/sanLimbu/eda-go/internal/system"
	"github.com/sanLimbu/eda-go/internal/web"
)

type monolith struct {
	*system.System
	modules []system.Module
}

func main() {
	if err := run(); err != nil {
		fmt.Printf("mallbots exitted abnormally: %s\n", err.Error())
		os.Exit(1)
	}
}

func run() (err error) {

	var cfg config.AppConfig
	cfg, err = config.InitConfig()
	if err != nil {
		return err
	}

	s, err := system.NewSystem(cfg)
	if err != nil {
		return err
	}
	m := monolith{
		System: s,
		modules: []system.Module{
			&baskets.Module{},
			&customers.Module{},
		},
	}

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			return
		}

	}(m.DB())

	err = m.MigrateDB(migrations.FS)
	if err != nil {
		return err
	}
	if err = m.startupModules(); err != nil {
		return err
	}

	//Mount general web resources
	m.Mux().Mount("/", http.FileServer(http.FS(web.WebUI)))

	fmt.Println("started mallbots apppliation")
	defer fmt.Println("stopped mallbots application")

	m.Waiter().Add(
		m.WaitForWeb,
		m.WaitForRPC,
		m.WaitForStream,
	)

	return m.Waiter().Wait()

}

func (m *monolith) startupModules() error {
	for _, module := range m.modules {
		ctx := m.Waiter().Context()
		if err := module.Startup(ctx, m); err != nil {
			return err
		}
	}

	return nil
}
