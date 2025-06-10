package catalog

import (
	"context"
	"encoding/json"

	elastic "gopkg.in/olivere/elastic.v5"
)

type Repository interface {
	Close()
	PutProduct(ctx context.Context, p Product) error
	GetProductByID(ctx context.Context, id string) (*Product, error)
	ListProducts(ctx context.Context, skip uint64, take uint64) ([]Product, error)
	ListProductsWithIDs(ctx context.Context, ids []string, skip uint64, take uint64) ([]Product, error)
	SearchProducts(ctx context.Context, query string, skip uint64, take uint64) ([]Product, error)
}

type elasticRepository struct {
	client *elastic.Client
}

type productDocument struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

func NewElasticRepository(url string) (Repository, error) {
	client, err := elastic.NewClient(elastic.SetURL(url), elastic.SetSniff(false))
	if err != nil {
		return nil, err
	}

	return &elasticRepository{client}, nil
}

func (r *elasticRepository) Close() {
	// r.client.CloseIndex("catalog")
}

func (r *elasticRepository) PutProduct(ctx context.Context, p Product) error {
	// Starts a new indexing request using the elastic search client
	_, err := r.client.Index().
	// Puts the product in the catalog index
		Index("catalog").
		// Type Product
		Type("product").
		// Sets the document ID as product ID
		Id(p.ID).
		// Converts Product struct to productDocument struct and then to BodyJson for elasticsearch
		BodyJson(productDocument{
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
		}).
		// Executes the request
		Do(ctx)
	return err

}

func (r *elasticRepository) GetProductByID(ctx context.Context, id string) (*Product, error) {
	// Get the data from the CATALOG index of type product with the provided ID
	res, err := r.client.Get().Index("catalog").Type("product").Id(id).Do(ctx)
	if err != nil {
		return nil, err
	}

	// Checks if the product was found
	if !res.Found {
		return nil, err
	}

	// Unmarshal the res.Source from elasticSearch and store in p
	p := productDocument{}
	if err = json.Unmarshal(*res.Source, &p); err != nil {
		return nil, err
	}
	return &Product{
		ID:          id,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
	}, err

}

func (r *elasticRepository) ListProducts(ctx context.Context, skip uint64, take uint64) ([]Product, error) {
	// Searches in the catalog index, of product type, and runs the query to match all, from skip to take
	res, err := r.client.Search().Index("catalog").Type("product").Query(elastic.NewMatchAllQuery()).From(int(skip)).Size(int(take)).Do(ctx)
	if err != nil {
		return nil, err
	}

	// Empty slice to store products
	products := []Product{}

	// range over the hits and store in the above slice
	for _, hit := range res.Hits.Hits {
		p := productDocument{}
		if err = json.Unmarshal(*hit.Source, &p); err == nil {
			products = append(products, Product{
				ID:          hit.Id,
				Name:        p.Name,
				Description: p.Description,
				Price:       p.Price,
			})
		}
	}
	return products, nil
}

func (r *elasticRepository) ListProductsWithIDs(ctx context.Context, ids []string, skip uint64, take uint64) ([]Product, error) {
	// Create a slice of elastic.MultiGetItem type
	items := []*elastic.MultiGetItem{}

	// Range over the productIDs
	for _, id := range ids {
		// Get Multiple items from catalog index of type product
		items = append(items, elastic.NewMultiGetItem().Index("catalog").Type("product").Id(id))
	}

	// Sends multi get request to elastic search
	res, err := r.client.MultiGet().Add(items...).Do(ctx)
	if err != nil {
		return nil, err
	}

	// Create empty slice for products
	products := []Product{}

	// range over the docs and append in the above slice
	for _, doc := range res.Docs {
		p := productDocument{}
		if err = json.Unmarshal(*doc.Source, &p); err == nil {
			products = append(products, Product{
				ID:          doc.Id,
				Name:        p.Name,
				Description: p.Description,
				Price:       p.Price,
			})
		}
	}
	return products, nil
}

func (r *elasticRepository) SearchProducts(ctx context.Context, query string, skip uint64, take uint64) ([]Product, error) {
	// Searches in the index catalog of type product with the given query from skip to take
	res, err := r.client.Search().Index("catalog").Type("products").Query(elastic.NewMultiMatchQuery(query, "name", "description")).From(int(skip)).Size(int(take)).Do(ctx)
	if err != nil {
		return nil, err
	}
	products := []Product{}
	for _, hits := range res.Hits.Hits {
		p := productDocument{}
		if err = json.Unmarshal(*hits.Source, &p); err == nil {
			products = append(products, Product{
				ID:          hits.Id,
				Name:        p.Name,
				Description: p.Description,
				Price:       p.Price,
			})
		}
	}
	return products, nil
}
