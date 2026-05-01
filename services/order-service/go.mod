module github.com/jason/ecommerce/order-service

go 1.26.2

require (
	github.com/go-sql-driver/mysql v1.8.1
	github.com/golang-migrate/migrate/v4 v4.17.1
	github.com/google/uuid v1.6.0
	github.com/jason/ecommerce/proto v0.0.0
	github.com/jmoiron/sqlx v1.4.0
	github.com/rabbitmq/amqp091-go v1.11.0
	google.golang.org/grpc v1.64.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	golang.org/x/net v0.52.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240318140521-94a12d6c2237 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
)

replace github.com/jason/ecommerce/proto => ../../proto
