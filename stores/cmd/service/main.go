package service

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/sanLimbu/eda-go/internal/config"
	"github.com/sanLimbu/eda-go/internal/system"
	"github.com/sanLimbu/eda-go/internal/web"
	"github.com/sanLimbu/eda-go/stores/migrations"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("stores exitted abnormally; %s\n", err)
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

	defer func(db *sql.DB) {
		if err = db.Close(); err != nil {
			return
		}
	}(s.DB())
	if err = s.MigrateDB(migrations.FS); err != nil {
		return err
	}

	s.Mux().Mount("/", http.FileServer(http.FS(web.WebUI)))
	// call the module composition root
	//if err = stores.

	fmt.Println("started stores service")
	defer fmt.Println("stopped stores service")

	s.Waiter().Add(
		s.WaitForWeb,
		s.WaitForRPC,
		s.WaitForStream,
	)

	return s.Waiter().Wait()

}
