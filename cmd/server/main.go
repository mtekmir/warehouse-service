package main

import (
	"log"

	"github.com/mtekmir/warehouse-service/internal/article"
	"github.com/mtekmir/warehouse-service/internal/config"
	"github.com/mtekmir/warehouse-service/internal/postgres"
	"github.com/mtekmir/warehouse-service/internal/product"
	"github.com/mtekmir/warehouse-service/internal/server"
)

func main() {
	if err := program(); err != nil {
		log.Fatal(err.Error())
	}
}

func program() error {
	c, err := config.Parse()
	if err != nil {
		return err
	}

	db, dbTidy, err := postgres.Setup(c.DBURL, c.DBMigrationsPath)
	if err != nil {
		return err
	}
	defer dbTidy()

	pr := postgres.NewProductRepo()
	ar := postgres.NewArticleRepo()
	
	ps := product.NewService(db, pr, ar)
	as := article.NewService(db, ar)

	s := server.NewServer(ps, as)

	if err := s.Start(); err != nil {
		return err
	}

	return nil
}
