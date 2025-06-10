package order

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/PranavTrip/go-grpc-graphql-ms/order/pb"
	"google.golang.org/grpc"
)

type Client struct {
	conn    *grpc.ClientConn
	service pb.OrderServiceClient
}

func NewClient(url string) (*Client, error) {

	// Checks if url is not empty
	if url == "" {
		return nil, fmt.Errorf("grpc target url cannot be empty")
	}
	// Creates a connection using grpc.Dial()
	conn, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	// Creates a new client
	c := pb.NewOrderServiceClient(conn)
	return &Client{conn, c}, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) PostOrder(
	ctx context.Context,
	accountID string,
	products []OrderedProduct,
) (*Order, error) {

	// Creates an empty slice to store the products in the protobuf format
	protoProducts := []*pb.PostOrderRequest_OrderProduct{}

	// Range over the products and append them to the protoProducts in the protobuf form
	for _, p := range products {
		protoProducts = append(protoProducts, &pb.PostOrderRequest_OrderProduct{
			ProductId: p.ID,
			Quantity:  p.Quantity,
		})
	}

	// Calls the PostOrder function to create an order
	r, err := c.service.PostOrder(
		ctx,
		&pb.PostOrderRequest{
			AccountId: accountID,
			Products:  protoProducts,
		},
	)
	if err != nil {
		return nil, err
	}

	// New order that is created
	newOrder := r.Order

	// Created at time for the new order
	newOrderCreatedAt := time.Time{}

	// Convert back to Time format from Binary
	newOrderCreatedAt.UnmarshalBinary(newOrder.CreatedAt)

	// return the order
	return &Order{
		ID:         newOrder.Id,
		CreatedAt:  newOrderCreatedAt,
		TotalPrice: newOrder.TotalPrice,
		AccountID:  newOrder.AccountId,
		Products:   products,
	}, nil
}

func (c *Client) GetOrdersForAccount(ctx context.Context, accountID string) ([]Order, error) {

	// Calls the function to Get orders for an account
	r, err := c.service.GetOrdersForAccount(ctx, &pb.GetOrdersForAccountRequest{
		AccountId: accountID,
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// create an empty slice to store the orders
	orders := []Order{}

	// Range over the orders and convert created_at from binary to time
	for _, orderProto := range r.Orders {
		newOrder := Order{
			ID:         orderProto.Id,
			TotalPrice: orderProto.TotalPrice,
			AccountID:  orderProto.AccountId,
		}
		newOrder.CreatedAt = time.Time{}
		newOrder.CreatedAt.UnmarshalBinary(orderProto.CreatedAt)

		// Empty slice to store products in the order
		products := []OrderedProduct{}
		
		// Range over the products and append them in the slice above
		for _, p := range orderProto.Products {
			products = append(products, OrderedProduct{
				ID:          p.Id,
				Quantity:    p.Quantity,
				Name:        p.Name,
				Description: p.Description,
				Price:       p.Price,
			})
		}
		newOrder.Products = products

		orders = append(orders, newOrder)
	}
	return orders, nil
}
