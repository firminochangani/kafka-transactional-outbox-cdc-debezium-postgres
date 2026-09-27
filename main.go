package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/firminochangani/kafka-cdc-debezium-postgres/inventorysvc"
	"github.com/firminochangani/kafka-cdc-debezium-postgres/ordersvc"
	"github.com/urfave/cli/v3"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cmd := &cli.Command{
		Name:        "ecommerce",
		Description: "Transactional Outbox implementation using CDC with Debezium and Postgres",
		Commands: []*cli.Command{
			ordersvc.RunOrderCmd(),
			inventorysvc.RunInventoryCmd(),
		},
	}
	err := cmd.Run(ctx, os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
