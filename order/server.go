package order

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"

	account "github.com/PranavTrip/go-grpc-graphql-ms/account"
	catalog "github.com/PranavTrip/go-grpc-graphql-ms/catalog"
	"github.com/PranavTrip/go-grpc-graphql-ms/order/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type grpcServer struct {
	pb.UnimplementedOrderServiceServer
	service       Service
	accountClient *account.Client
	catalogClient *catalog.Client
}

func ListenGRPC(s Service, accountURL string, catalogURL string, port int) error {
	accountClient, err := account.NewClient(accountURL)
	if err != nil {
		return err
	}
	catalogClient, err := catalog.NewClient(catalogURL)
	if err != nil {
		accountClient.Close()
		return err
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		accountClient.Close()
		catalogClient.Close()
		return err
	}

	serv := grpc.NewServer()
	pb.RegisterOrderServiceServer(serv, &grpcServer{
		UnimplementedOrderServiceServer: pb.UnimplementedOrderServiceServer{},
		service:                         s,
		accountClient:                   accountClient,
		catalogClient:                   catalogClient,
	})
	reflection.Register(serv)
	return serv.Serve(lis)

}

func (s *grpcServer) PostOrder(ctx context.Context, r *pb.PostOrderRequest) (*pb.PostOrderResponse, error) {

	// Get account from account client using the accountID
	_, err := s.accountClient.GetAccount(ctx, r.AccountId)
	if err != nil {
		log.Println("Error getting account: ", err)
		return nil, errors.New("account not found")
	}

	// Empty slice for storing the product IDs of Ordered Products
	productIDs := []string{}

	// Range over the products coming from request and append in the above slice
	for _, p := range r.Products {
		productIDs = append(productIDs, p.ProductId)
	}

	// Now based on the productIDs of the ordered products, get the entire products using the catalogClient
	orderedProducts, err := s.catalogClient.GetProducts(ctx, productIDs, 0, 0, "")
	if err != nil {
		log.Println("Error getting products: ", err)
		return nil, errors.New("products not found")
	}

	// Create empty slice for storing the ordered products
	products := []OrderedProduct{}

	// Range over the orderedProducts to include only the products with valid quantity
	for _, p := range orderedProducts {
		product := OrderedProduct{
			ID:          p.ID,
			Quantity:    0,
			Price:       p.Price,
			Name:        p.Name,
			Description: p.Description,
		}
		for _, rp := range r.Products {
			if rp.ProductId == p.ID {
				product.Quantity = rp.Quantity
				break
			}
		}

		if product.Quantity != 0 {
			products = append(products, product)
		}
	}

	// Call the service function to post the order in the DB
	order, err := s.service.PostOrder(ctx, r.AccountId, products)
	if err != nil {
		log.Println("Error posting order: ", err)
		return nil, errors.New("could not post order")
	}

	// Convert the order to protobuf to match the return statements
	orderProto := &pb.Order{
		Id:         order.ID,
		AccountId:  order.AccountID,
		TotalPrice: order.TotalPrice,
		Products:   []*pb.Order_OrderProduct{},
	}
	orderProto.CreatedAt, _ = order.CreatedAt.MarshalBinary()
	for _, p := range order.Products {
		orderProto.Products = append(orderProto.Products, &pb.Order_OrderProduct{
			Id:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Quantity:    p.Quantity,
		})
	}
	return &pb.PostOrderResponse{
		Order: orderProto,
	}, nil
}

func (s *grpcServer) GetOrdersForAccount(ctx context.Context, r *pb.GetOrdersForAccountRequest) (*pb.GetOrdersForAccountResponse, error) {

	// Call the service function to Get all orders for a particular accountID
	accountOrders, err := s.service.GetOrderForAccount(ctx, r.AccountId)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Create a map to store all the product IDs in the order
	productIDMap := map[string]bool{}

	// Range over the orders of a particular account to store the product IDs avoiding duplicates - using map
	for _, o := range accountOrders {
		for _, p := range o.Products {
			productIDMap[p.ID] = true
		}
	}

	// Creating an empty slice to store all productIDs in the map
	productIDs := []string{}
	// Range over the map to fill the above slice
	for id := range productIDMap {
		productIDs = append(productIDs, id)
	}

	// Call the GetProducts from catalogClient to get all products using the productIDs
	products, err := s.catalogClient.GetProducts(ctx, productIDs, 0, 0, "")
	if err != nil {
		log.Println("Error getting account products: ", err)
		return nil, err
	}

	// Constructing a variable to store the orders in the protobuf response
	orders := []*pb.Order{}

	// Range over the account orders to return the values
	for _, o := range accountOrders {
		op := &pb.Order{
			AccountId:  o.AccountID,
			Id:         o.ID,
			TotalPrice: o.TotalPrice,
			Products:   []*pb.Order_OrderProduct{},
		}
		// Marshalling time to send over grpc
		op.CreatedAt, _ = o.CreatedAt.MarshalBinary()

		// Range over o.Products (it has only ID and quantity)
		for _, product := range o.Products {
			// range over products (it has all the product info)
			for _, p := range products {
				if p.ID == product.ID {
					product.Name = p.Name
					product.Description = p.Description
					product.Price = p.Price
					break
				}
			}
			// Convert to grpc format
			op.Products = append(op.Products, &pb.Order_OrderProduct{
				Id:          product.ID,
				Name:        product.Name,
				Description: product.Description,
				Price:       product.Price,
				Quantity:    product.Quantity,
			})
		}

		orders = append(orders, op)
	}
	return &pb.GetOrdersForAccountResponse{Orders: orders}, nil
}
