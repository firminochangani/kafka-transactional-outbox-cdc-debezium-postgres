package inventorysvc

import (
	"context"

	"github.com/urfave/cli/v3"
)

func RunInventoryCmd() *cli.Command {
	return &cli.Command{
		Name:        "inventory-svc",
		Description: "Inventory service is responsible for consuming the event OrderPlaced",
		Action:      run,
	}
}

func run(ctx context.Context, command *cli.Command) error {

	// consume events from work.firmino.orders topic

	return nil
}
