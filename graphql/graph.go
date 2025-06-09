package main

import (
	"github.com/99designs/gqlgen/graphql"
	"github.com/PranavTrip/go-grpc-graphql-ms/account"
	"github.com/PranavTrip/go-grpc-graphql-ms/catalog"
	"github.com/PranavTrip/go-grpc-graphql-ms/order"
)

type Server struct {
	accountClient *account.Client
	catalogClient *catalog.Client
	orderClient   *order.Client
}

func NewGraphQLServer(accountURL, catalogURL, orderURL string) (*Server, error) {
	accountClient, err := account.NewClient(accountURL)
	if err != nil {
		return nil, err
	}

	catalogClient, err := catalog.NewClient(catalogURL)
	if err != nil {
		accountClient.Close()
		return nil, err
	}

	orderClient, err := order.NewClient(orderURL)

	if err != nil {
		accountClient.Close()
		catalogClient.Close()
		return nil, err
	}

	return &Server{
		accountClient, catalogClient, orderClient,
	}, nil
}

func (s *Server) Mutation() MutationResolver {
	return &mutationResolver{
		server: s,
	}
}

func (s *Server) Query() QueryResolver {
	return &queryResolver{
		server: s,
	}
}

func (s *Server) Account() AccountResolver {
	return &accountResolver{
		server: s,
	}
}

func (s *Server) ToExecutableSchema() graphql.ExecutableSchema {
	return NewExecutableSchema(Config{
		Resolvers: s,
	})
}
