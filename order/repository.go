package order

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

type Repository interface {
	Close()
	PutOrder(ctx context.Context, o Order) error
	GetOrderForAccount(ctx context.Context, accountID string) ([]Order, error)
}

type postgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(url string) (Repository, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return &postgresRepository{db}, nil
}

func (r *postgresRepository) Close() {
	r.db.Close()
}

func (r *postgresRepository) PutOrder(ctx context.Context, o Order) (err error) {
	// Begin a transaction to post an order
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// A function to rollback the transaction in case of any error; otherwise commit the transaction
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	// ExexContext to execute the SQL command
	_, err = tx.ExecContext(ctx, "INSERT INTO orders(id, created_at, account_id, total_price) VALUES($1, $2, $3, $4)", o.ID, o.CreatedAt, o.AccountID, o.TotalPrice)
	if err != nil {
		return
	}

	// Prepare context to put products in the order
	stmt, err := tx.PrepareContext(ctx, pq.CopyIn("order_products", "order_id", "product_id", "quantity"))
	if err != nil {
		return
	}

	// Range over the o.Products to put the products in the order based on the order ID
	for _, p := range o.Products {
		_, err = stmt.ExecContext(ctx, o.ID, p.ID, p.Quantity)
		if err != nil {
			return

		}
	}

	// ExecContext to put the products
	_, err = stmt.ExecContext(ctx)
	if err != nil {
		return
	}
	stmt.Close()
	return
}

func (r *postgresRepository) GetOrderForAccount(ctx context.Context, accountID string) ([]Order, error) {

	// QueryContext to get data from the DB
	rows, err := r.db.QueryContext(ctx, `
		SELECT o.id, o.created_at, o.account_id, o.total_price::money::numeric::float8, op.product_id, op.quantity
		FROM orders o JOIN order_products op ON (o.id = op.order_id)
		WHERE account_id = $1
		ORDER BY o.id
		`, accountID,
	)

	if err != nil {
		return nil, err
	}

	// Close the rows
	defer rows.Close()

	// Orders slice - to be returned as the function response
	orders := []Order{}

	// Keeps track of current order
	order := &Order{}

	// Keeps track of previous order
	lastOrder := &Order{}

	// The ordered product
	orderedProduct := &OrderedProduct{}

	// List of products in the current order
	products := []OrderedProduct{}

	// Loop over all the rows
	for rows.Next() {

		// Load the data of current row in the above defined variables
		var productID string
		var quantity int

		if err = rows.Scan(
			&order.ID,
			&order.CreatedAt,
			&order.AccountID,
			&order.TotalPrice,
			&orderedProduct.ID,
			&orderedProduct.Quantity,
		); err != nil {
			return nil, err
		}

		// Detects if we have a new order - appends the current order to orders slice
		if lastOrder.ID != "" && lastOrder.ID != order.ID {
			newOrder := Order{
				ID:         lastOrder.ID,
				CreatedAt:  lastOrder.CreatedAt,
				AccountID:  lastOrder.AccountID,
				TotalPrice: lastOrder.TotalPrice,
				Products:   lastOrder.Products,
			}
			orders = append(orders, newOrder)

			// Reset the list of products
			products = []OrderedProduct{}
		}
		// Add current product to current order
		products = append(products, OrderedProduct{
			ID:       orderedProduct.ID,
			Quantity: orderedProduct.Quantity,
		})
		// Setting last order as current order
		*lastOrder = *order
	}

	// Add the last order to the orders slice
	if lastOrder != nil {
		newOrder := Order{
			ID:         lastOrder.ID,
			AccountID:  lastOrder.AccountID,
			CreatedAt:  lastOrder.CreatedAt,
			TotalPrice: lastOrder.TotalPrice,
			Products:   lastOrder.Products,
		}
		orders = append(orders, newOrder)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil

}
