cd generate
go run main.go -dns "username:password@tcp(localhost:3306)/database?charset=utf8mb4&parseTime=True&loc=Local" -out ../demo -mod github.com/hellobchain/nacos-go-demo