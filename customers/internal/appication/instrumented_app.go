package appication

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
)

type instrumentApp struct {
	App
	customerRgistered prometheus.Counter
}

var _ App = (*instrumentApp)(nil)

func NewInstrumentedApp(app App, customersRegistered prometheus.Counter) App {
	return instrumentApp{
		App:               app,
		customerRgistered: customersRegistered,
	}
}

func (a instrumentApp) RegisterCustomer(ctx context.Context, register RegisterCustomer) error {
	err := a.App.RegisterCustomer(ctx, register)
	if err != nil {
		return err
	}
	a.customerRgistered.Inc()
	return nil
}
