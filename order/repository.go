package order

import (
	"context"
	"database/sql"
	"time"

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
    defer rows.Close()

    orders := []Order{}
    var currentOrder *Order = nil
    var currentProducts []OrderedProduct = nil

    for rows.Next() {
        var orderID string
        var createdAt time.Time
        var accountIDFromDB string
        var totalPrice float64
        var rawProductID sql.RawBytes
        var quantity uint32

        // Scan into local variables
        err = rows.Scan(
            &orderID,
            &createdAt,
            &accountIDFromDB,
            &totalPrice,
            &rawProductID, 
            &quantity,
        )
        if err != nil {
            return nil, err
        }

        // Convert rawProductID to string after successful scan
        productID := string(rawProductID)

        // Detect if we have a new order
        if currentOrder == nil || currentOrder.ID != orderID {
            if currentOrder != nil {
                currentOrder.Products = currentProducts
                orders = append(orders, *currentOrder)
            }
            currentOrder = &Order{
                ID:         orderID,
                CreatedAt:  createdAt,
                AccountID:  accountIDFromDB,
                TotalPrice: totalPrice,
            }
            currentProducts = []OrderedProduct{}
        }

        currentProducts = append(currentProducts, OrderedProduct{
            ID:       productID,
            Quantity: quantity,
        })
    }

    if currentOrder != nil {
        currentOrder.Products = currentProducts
        orders = append(orders, *currentOrder)
    }

    if err = rows.Err(); err != nil {
        return nil, err
    }
    return orders, nil
}