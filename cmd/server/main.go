package main

import (
	"log"
	_ "net/http/pprof"

	"github.com/mtekmir/warehouse-service/internal/article"
	"github.com/mtekmir/warehouse-service/internal/config"
	"github.com/mtekmir/warehouse-service/internal/logs"
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

	logger, err := logs.NewLogger(&c.Env, c.LogFile)
	if err != nil {
		return err
	}

	db, dbTidy, err := postgres.Setup(logger, c.DBURL, c.DBMigrationsPath)
	if err != nil {
		return err
	}
	defer dbTidy()

	pr := postgres.NewProductRepo()
	ar := postgres.NewArticleRepo()

	ps := product.NewService(logger, db, pr, ar)
	as := article.NewService(logger, db, ar)

	s := server.NewServer(logger, ps, as)

	if err := s.Start(c.Port, c.WriteTimeout, c.ReadTimeout, c.IdleTimeout); err != nil {
		return err
	}

	return nil
}
