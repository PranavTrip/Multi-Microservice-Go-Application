package account

import (
	"context"
	"fmt"

	"github.com/PranavTrip/go-grpc-graphql-ms/account/pb"
	"google.golang.org/grpc"
)

type Client struct {
	conn    *grpc.ClientConn
	service pb.AccountServiceClient
}

func NewClient(url string) (*Client, error) {
	if url == "" {
		return nil, fmt.Errorf("grpc target url cannot be empty")
	}
	conn, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	client := pb.NewAccountServiceClient(conn)
	return &Client{conn, client}, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) PostAccount(ctx context.Context, name string) (*Account, error) {
	res, err := c.service.PostAccount(ctx, &pb.PostAccountRequest{Name: name})
	if err != nil {
		return nil, err
	}

	return &Account{
		ID:   res.Account.Id,
		Name: res.Account.Name,
	}, nil
}

func (c *Client) GetAccount(ctx context.Context, id string) (*Account, error) {
	res, err := c.service.GetAccount(ctx, &pb.GetAccountRequest{Id: id})
	if err != nil {
		return nil, err
	}

	return &Account{
		ID:   res.Account.Id,
		Name: res.Account.Name,
	}, nil
}

func (c *Client) GetAccounts(ctx context.Context, skip uint64, take uint64) ([]Account, error) {
	res, err := c.service.GetAccounts(ctx, &pb.GetAccountsRequest{Skip: skip, Take: take})
	if err != nil {
		return nil, err
	}
	accounts := []Account{}
	for _, a := range res.Accounts {
		accounts = append(accounts, Account{
			ID:   a.Id,
			Name: a.Name,
		})
	}
	return accounts, nil

}
