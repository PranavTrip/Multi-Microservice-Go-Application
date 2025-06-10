package account

import (
	"context"
	"fmt"
	"net"

	"github.com/PranavTrip/go-grpc-graphql-ms/account/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type grpcServer struct {
	pb.UnimplementedAccountServiceServer
	service Service
}

func ListenGRPC(s Service, port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	serv := grpc.NewServer()
	pb.RegisterAccountServiceServer(serv, &grpcServer{UnimplementedAccountServiceServer: pb.UnimplementedAccountServiceServer{}, service: s})
	reflection.Register(serv)
	return serv.Serve(lis)
}

func (s *grpcServer) PostAccount(ctx context.Context, r *pb.PostAccountRequest) (*pb.PostAccountResponse, error) {
	// Call the service function to create account
	a, err := s.service.PostAccount(ctx, r.Name)
	if err != nil {
		return nil, err
	}
	return &pb.PostAccountResponse{
		Account: &pb.Account{
			Id:   a.ID,
			Name: a.Name,
		},
	}, nil
}

func (s *grpcServer) GetAccount(ctx context.Context, r *pb.GetAccountRequest) (*pb.GetAccountResponse, error) {
	// Call the service function to get the account with particular id
	acc, err := s.service.GetAccount(ctx, r.Id)
	if err != nil {
		return nil, err
	}

	// Convert the response to protobuf format for grpc
	return &pb.GetAccountResponse{
		Account: &pb.Account{
			Id:   acc.ID,
			Name: acc.Name,
		},
	}, err
}

func (s *grpcServer) GetAccounts(ctx context.Context, r *pb.GetAccountsRequest) (*pb.GetAccountsResponse, error) {
	// Call the service function to Get all accounts
	res, err := s.service.GetAccounts(ctx, r.Skip, r.Take)
	if err != nil {
		return nil, err
	}

	// Empty slice to store all accounts
	accounts := []*pb.Account{}

	// Range over the response and fill in the slice
	for _, p := range res {
		accounts = append(accounts, &pb.Account{
			Id:   p.ID,
			Name: p.Name,
		})
	}
	return &pb.GetAccountsResponse{Accounts: accounts}, nil
}
