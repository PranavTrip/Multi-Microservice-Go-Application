package catalog

import (
	"context"

	"github.com/PranavTrip/go-grpc-graphql-ms/catalog/pb"
	"google.golang.org/grpc"
)

type Client struct {
	conn    *grpc.ClientConn
	service pb.CatalogServiceClient
}

func NewClient(url string) (*Client, error) {
	conn, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	c := pb.NewCatalogServiceClient(conn)
	return &Client{conn, c}, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) PostProduct(ctx context.Context, name string, description string, price float64) (*Product, error) {
	res, err := c.service.PostProduct(ctx, &pb.PostProductRequest{
		Name:        name,
		Description: description,
		Price:       price,
	})
	if err != nil {
		return nil, err
	}
	return &Product{
		ID:          res.Product.Id,
		Name:        res.Product.Name,
		Description: res.Product.Description,
		Price:       res.Product.Price,
	}, nil

}

func (c *Client) GetProduct(ctx context.Context, id string) (*Product, error) {
	res, err := c.service.GetProduct(ctx, &pb.GetProductRequest{
		Id: id,
	})
	if err != nil {
		return nil, err
	}
	return &Product{
		ID:          res.Product.Id,
		Name:        res.Product.Name,
		Description: res.Product.Description,
		Price:       res.Product.Price,
	}, nil
}

func (c *Client) GetProducts(ctx context.Context, ids []string, skip uint64, take uint64, query string) ([]Product, error) {
	res, err := c.service.GetProducts(ctx, &pb.GetProductsRequest{
		Query: query,
		Ids:   ids,
		Skip:  skip,
		Take:  take,
	})
	if err != nil {
		return nil, err
	}
	products := []Product{}
	for _, r := range res.Products {
		products = append(products, Product{
			ID:          r.Id,
			Name:        r.Name,
			Description: r.Description,
			Price:       r.Price,
		})
	}
	return products, nil
}
