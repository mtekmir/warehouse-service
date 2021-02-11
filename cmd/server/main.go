package main

import (
	"log"

	"github.com/mtekmir/warehouse-service/internal/config"
	"github.com/mtekmir/warehouse-service/internal/postgres"
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

	_, dbTidy, err := postgres.Setup(c.DBURL, c.DBMigrationsPath)
	if err != nil {
		return err
	}
	defer dbTidy()

	return nil
}
