package baskets

import (
	"context"

	"github.com/go-openapi/runtime"
)

type Client interface {
	StartBasket(ctx context.Context, customerID string) (string, error)
	CheckoutBasket(ctx context.Context, basketID, paymentID string) error
	CancelBasket(ctx context.Context, basketID string) error
	AddItem(ctx context.Context, basketID, productID string, quantity int) error
}

type client struct {
	c
}

func NewClient(transport runtime.ClientTransport) Client {
	return &Client
}
