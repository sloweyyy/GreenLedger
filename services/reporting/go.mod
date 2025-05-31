module github.com/sloweyyy/GreenLedger/services/reporting

go 1.23

toolchain go1.24.2

require (
	github.com/google/uuid v1.3.1
	github.com/jung-kurt/gofpdf v1.16.2
	github.com/shopspring/decimal v1.3.1
	github.com/sloweyyy/GreenLedger/shared v0.0.0
	github.com/wcharczuk/go-chart/v2 v2.1.1
	gorm.io/gorm v1.25.5
)

require (
	github.com/blend/go-sdk v1.20240719.1 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.4.3 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/image v0.11.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	gorm.io/driver/postgres v1.5.3 // indirect
)

replace github.com/sloweyyy/GreenLedger/shared => ../../shared
