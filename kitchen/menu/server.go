package main

import (
	"log"
	"makoskitchen/kitchen/menu/graph"
	"makoskitchen/kitchen/menu/graph/generated"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

const defaultPort = "9080"

func main() {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = defaultPort
	}
	db_user := os.Getenv("DB_USER")
	db_pass := os.Getenv("DB_PASS")
	db_host := os.Getenv("DB_HOST")
	db_port := os.Getenv("DB_PORT")
	db_name := os.Getenv("DB_NAME")
	db_svc := os.Getenv("DB_SERVICE")

	dataSource := db_user + ":" + db_pass + "@tcp(" + db_host + ":" + db_port + ")/" + db_name + "?charset=utf8&parseTime=True&loc=Local"

	db, err := gorm.Open(db_svc, dataSource)
	if err != nil {
		panic(err)
	}
	if db == nil {
		panic(err)
	}
	defer func() {
		if db != nil {
			if err := db.Close(); err != nil {
				panic(err)
			}
		}
	}()
	db.LogMode(true)

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{DB: db}}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
