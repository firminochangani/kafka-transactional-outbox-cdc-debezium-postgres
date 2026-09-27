package ordersvc

import (
	"context"

	"github.com/urfave/cli/v3"
)

func RunOrderCmd() *cli.Command {
	return &cli.Command{
		Name:        "order-svc",
		Description: "Order service is responsible for storing the order details and producing OrderPlaced transactionally",
		Action:      run,
	}
}

func run(ctx context.Context, command *cli.Command) error {

	// expose an endpoint to place orders;
	// store order in the orders table and OrderPlaced events in outbox table;

	return nil
}
