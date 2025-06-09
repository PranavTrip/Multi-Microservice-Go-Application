package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/handler"
	"github.com/kelseyhightower/envconfig"
)

type AppConfig struct {
	AccountURL string `envconfig:"ACCOUNT_SERVICE_URL"`
	CatalogURL string `envconfig:"CATALOG_SERVICE_URL"`
	OrderURL   string `envconfig:"ORDER_SERVICE_URL"`
}

func main() {
	fmt.Println("ACCOUNT_SERVICE_URL =", os.Getenv("ACCOUNT_SERVICE_URL"))
	fmt.Println("CATALOG_SERVICE_URL =", os.Getenv("CATALOG_SERVICE_URL"))
	fmt.Println("ORDER_SERVICE_URL =", os.Getenv("ORDER_SERVICE_URL"))
	var cfg AppConfig
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}
	s, err := NewGraphQLServer(cfg.AccountURL, cfg.CatalogURL, cfg.OrderURL)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/graphql",handler.GraphQL(s.ToExecutableSchema()))
	http.Handle("/playground", handler.Playground("pranav","/graphql"))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
