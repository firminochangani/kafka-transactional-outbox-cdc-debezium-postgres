package ordersvc

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	_ "github.com/lib/pq"
	"github.com/oklog/ulid/v2"
	"github.com/urfave/cli/v3"

	"github.com/firminochangani/kafka-cdc-debezium-postgres/events"
)

func RunOrderCmd() *cli.Command {
	return &cli.Command{
		Name:        "order-svc",
		Description: "Order service is responsible for storing the order details and producing OrderPlaced transactionally",
		Action:      run,
	}
}

func run(ctx context.Context, command *cli.Command) error {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	err = db.PingContext(ctx)
	if err != nil {
		return err
	}

	handlers := &Handlers{db: db}
	router := echo.New()
	router.Use(middleware.Recover())
	router.Use(middleware.RequestLogger())

	router.POST("/orders", handlers.PlaceOrder)

	return router.Start(":8080")
}

type Handlers struct {
	db *sql.DB
}

func (h *Handlers) PlaceOrder(c *echo.Context) error {
	ctx := c.Request().Context()
	var body PlaceOrderRequest
	if err := c.Bind(&body); err != nil {
		return err
	}

	order := body.Order
	order.ID = ulid.Make().String()
	order.PlacedAt = time.Now()

	// ACID compliant
	err := Transact(ctx, h.db, func(tx *sql.Tx) error {
		eventPublisher := &EventPublisher{tx}
		orderRepository := &OrderRepository{tx}

		err := orderRepository.Save(ctx, order)
		if err != nil {
			return err
		}

		return eventPublisher.PublishOrderPlaced(ctx, mapOrderToEvent(order))
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, order)
}

//
// Models
//

type PlaceOrderRequest struct {
	Order Order `json:"order"`
}

type Order struct {
	ID         string      `json:"id"`
	CustomerID string      `json:"customer_id"`
	Items      []OrderItem `json:"items"`
	PlacedAt   time.Time   `json:"placed_at"`
}
type OrderItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Name      string `json:"name"`
	Price     int    `json:"price"`
}

// Mappers
func mapOrderToEvent(order Order) events.OrderPlaced {
	items := make([]events.OrderPlacedItem, len(order.Items))
	for i, item := range order.Items {
		items[i] = events.OrderPlacedItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Name:      item.Name,
			Price:     item.Price,
		}
	}

	return events.OrderPlaced{
		EventHeader: events.EventHeader{
			CorrelationID: "", // TODO: out of scope
			Service:       "orders-svc",
		},
		OrderID:  order.ID,
		Items:    items,
		PlacedAt: order.PlacedAt,
	}
}

type DBExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// OrderRepository Abstracts away SQL operations related to orders from the caller;
type OrderRepository struct {
	db DBExecutor
}

func (h *OrderRepository) Save(ctx context.Context, order Order) error {
	_, err := h.db.ExecContext(
		ctx,
		`INSERT INTO orders_svc.orders (id, customer_id, placed_at) VALUES ($1, $2, $3)`,
		order.ID,
		order.CustomerID,
		order.PlacedAt,
	)
	if err != nil {
		return fmt.Errorf("order-repository: error saving order: %w", err)
	}

	return h.saveOrderItems(ctx, order)
}

func (h *OrderRepository) saveOrderItems(ctx context.Context, order Order) error {
	for _, item := range order.Items {
		_, err := h.db.ExecContext(
			ctx,
			`INSERT INTO orders_svc.order_items (id, product_id, order_id, name, price, quantity) VALUES ($1, $2, $3, $4, $5, $6)`,
			ulid.Make().String(),
			item.ProductID,
			order.ID,
			item.Name,
			item.Price,
			item.Quantity,
		)
		if err != nil {
			return fmt.Errorf("order-repository: error saving order item: %w", err)
		}
	}

	return nil
}

// EventPublisher Publishes events to the outbox table. Abstracts away SQL operations from the caller;
type EventPublisher struct {
	db DBExecutor
}

func (h *EventPublisher) PublishOrderPlaced(ctx context.Context, event events.OrderPlaced) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("eventbus: error marshalling event: %w", err)
	}

	_, err = h.db.ExecContext(
		ctx,
		`INSERT INTO orders_svc.outbox (id, aggregateid, aggregatetype, type, payload) VALUES(nextval('orders_svc.outbox_id_seq'::regclass), $1, 'orders', 'OrderPlaced', $2)`,
		event.OrderID,
		payload,
	)
	if err != nil {
		return fmt.Errorf("eventbus: error saving event into the outbox table: %w", err)
	}

	return nil
}

// Transact Transaction script
func Transact(ctx context.Context, db *sql.DB, handler func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	err = handler(tx)
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			return errors.Join(err, rollbackErr)
		}

		return err
	}

	return tx.Commit()
}
