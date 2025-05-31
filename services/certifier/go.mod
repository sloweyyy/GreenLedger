module github.com/greenledger/services/certifier

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/google/uuid v1.3.1
	github.com/greenledger/shared v0.0.0
	github.com/lib/pq v1.10.9
	github.com/stretchr/testify v1.8.4
	github.com/swaggo/gin-swagger v1.6.0
	github.com/swaggo/swag v1.16.2
	google.golang.org/grpc v1.58.3
	google.golang.org/protobuf v1.31.0
	gorm.io/driver/postgres v1.5.3
	gorm.io/gorm v1.25.5
	github.com/shopspring/decimal v1.3.1
	github.com/ethereum/go-ethereum v1.13.4
	crypto/sha256
	encoding/hex
)

replace github.com/greenledger/shared => ../../shared
