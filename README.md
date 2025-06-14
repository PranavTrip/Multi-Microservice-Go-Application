## Multi-Microservice Go E-commerce Application

The project consists of the following main components:

- Account Service
- Catalog Service
- Order Service
- GraphQL API Gateway

Each service has its own database:
- Account and Order services use PostgreSQL
- Catalog service uses Elasticsearch

---

## Getting Started

1. Clone the repository:
   ```
   git clone <repository-url>
   cd <project-directory>
   ```

2. Start the services using Docker Compose:
   ```
   docker compose up
   ```

3. Access the GraphQL playground at `http://localhost:8000/playground`

---

## Tech Stack

- <img src="https://raw.githubusercontent.com/golang-samples/gopher-vector/master/gopher.png" alt="Go" width="20" height="20"/> **Go** – Core backend language for all microservices
- <img src="https://grpc.io/img/logos/grpc-icon-color.png" alt="gRPC" width="20" height="20"/> **gRPC** – Fast and type-safe service-to-service communication
- <img src="https://upload.wikimedia.org/wikipedia/commons/1/17/GraphQL_Logo.svg" alt="GraphQL" width="20" height="20"/> **GraphQL** – Centralized API gateway
- <img src="https://www.docker.com/wp-content/uploads/2022/03/Moby-logo.png" alt="Docker" width="20" height="20"/> **Docker** – Containerization of services
- <img src="https://www.docker.com/wp-content/uploads/2022/03/Moby-logo.png" alt="Docker Compose" width="20" height="20"/> **Docker Compose** – Local orchestration
- <img src="https://upload.wikimedia.org/wikipedia/commons/2/29/Postgresql_elephant.svg" alt="PostgreSQL" width="20" height="20"/> **PostgreSQL** – Relational data storage

---

## Command to Generate Protobuf Files

To generate Go code from your `.proto` files, run:

    protoc --go_out=./pb --go-grpc_out=./pb account.proto

### Install protoc
To install protoc:
```
brew install protoc
```

### Required Extensions

You need these plugins installed:

- `protoc-gen-go`
- `protoc-gen-go-grpc`

### How to Install

Run these commands:
```
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Make sure your Go binaries path is in your `PATH` environment variable. Add this line to your shell profile (`~/.bashrc` or `~/.zshrc`):
```
    export PATH="$PATH:$(go env GOPATH)/bin"
```

Reload your shell or run `source ~/.bashrc` (or your profile) after adding the line.

---

Now `protoc` can find the plugins and generate Go and gRPC code correctly.

---

## GraphQL API Usage

The GraphQL API provides a unified interface to interact with all the microservices.

### Query Accounts

```graphql
query {
  accounts {
    id
    name
  }
}
```

### Create an Account

```graphql
mutation {
  createAccount(account: {name: "New Account"}) {
    id
    name
  }
}
```

### Query Products

```graphql
query {
  products {
    id
    name
    price
  }
}
```

### Create a Product

```graphql
mutation {
  createProduct(product: {name: "New Product", description: "A new product", price: 19.99}) {
    id
    name
    price
  }
}
```

### Create an Order

```graphql
mutation {
  createOrder(order: {accountId: "account_id", products: [{id: "product_id", quantity: 2}]}) {
    id
    totalPrice
    products {
      name
      quantity
    }
  }
}
```

### Query Account with Orders

```graphql
query {
  accounts(id: "account_id") {
    name
    orders {
      id
      createdAt
      totalPrice
      products {
        name
        quantity
        price
      }
    }
  }
}
```

## Advanced Queries

### Pagination and Filtering

```graphql
query {
  products(pagination: {skip: 0, take: 5}, query: "search_term") {
    id
    name
    description
    price
  }
}
```

### Calculate Total Spent by an Account

```graphql
query {
  accounts(id: "account_id") {
    name
    orders {
      totalPrice
    }
  }
}
```

---

## Access Elastic Search for Catalog DB

After the app is up and running, ElasticSearch DB can be accessed at:
```
    http://localhost:9200
```

To check all indices visit:
```
http://localhost:9200/_cat/indices?v
```

To check all documents in the catalog index:
```
http://localhost:9200/catalog/_search?pretty
```

---

## Access Postgres for Account DB

PostgreSQL can be accessed and used via CLI (Command Line Interface)
After the app is up and running, in the terminal follow the below commands:

```
docker exec -it <container_id_for_accountDB> psql -U <db_username> -d <db_name>
```

Once you are in the PostgreSQL CLI, use the following commands:

To view tables:
```
 \dt
```

To view accounts information:
```
SELECT * FROM accounts;
````

---

## Access Postgres for Orders DB

PostgreSQL can be accessed and used via CLI (Command Line Interface)
After the app is up and running, in the terminal follow the below commands:

```
docker exec -it <container_id_for_orderDB> psql -U <db_username> -d <db_name>
```

Once you are in the PostgreSQL CLI, use the following commands:

To view tables:
```
 \dt
```

To view orders information:
```
SELECT * FROM order;
````

To view ordered products information:
```
SELECT * FROM order_products;
````
