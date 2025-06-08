package order

import (
	"context"
	"time"

	"github.com/segmentio/ksuid"
)

type Service interface {
	PostOrder(ctx context.Context, accountID string, products []OrderedProduct) (*Order, error)
	GetOrderForAccount(ctx context.Context, accountID string) ([]Order, error)
}

type Order struct {
	ID         string
	CreatedAt  time.Time
	TotalPrice float64
	AccountID  string
	Products   []OrderedProduct
}

type OrderedProduct struct {
	ID          string
	Name        string
	Description string
	Price       float64
	Quantity    uint32
}

type orderService struct {
	repository Repository
}

func NewService(r Repository) Service {
	return &orderService{r}
}

func (s *orderService) PostOrder(ctx context.Context, accountID string, products []OrderedProduct) (*Order, error) {
	order := &Order{
		ID:        ksuid.New().String(),
		CreatedAt: time.Now().UTC(),
		AccountID: accountID,
		Products:  products,
	}
	order.TotalPrice = 0.0
	for _, p := range order.Products {
		order.TotalPrice += p.Price * float64(p.Quantity)
	}
	err := s.repository.PutOrder(ctx, *order)
	if err != nil {
		return nil, err
	}
	return order, nil

}

func (s orderService) GetOrderForAccount(ctx context.Context, accountID string) ([]Order, error) {
	return s.repository.GetOrderForAccount(ctx, accountID)
}
