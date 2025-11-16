module test-project

go 1.21

require (
	go.uber.org/fx v1.20.0
	github.com/gorilla/mux v1.8.0
	gorm.io/gorm v1.25.5
	gorm.io/driver/sqlite v1.5.4
	github.com/golang-jwt/jwt/v5 v5.2.0

)

replace test-project => ./
