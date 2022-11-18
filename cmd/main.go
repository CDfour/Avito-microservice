package main

import (
	"context"
	"fmt"

	"Avito/internal/api"
	"Avito/internal/config"
	"Avito/internal/controller"
	"Avito/internal/repository"

	_ "Avito/docs"

	"github.com/gin-gonic/gin"
	pgx "github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title          Microservice for working with user balance
// @version        1.0

// @contact.name   Ilya
// @contact.email  biv_1998@mail.ru

// @host           localhost:8080
// @BasePath       /
func main() {

	config, err := config.LoadConfig()
	if err != nil {
		logrus.Errorln("Load config: ", err)
		panic(err)
	}

	dbURL := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", config.Username, config.Password, config.Host, config.Port, config.Database)
	db, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		logrus.Errorln("Connect to DB", err)
		panic(err)
	}
	defer db.Close(context.Background())

	repository, err := repository.NewRepository(db)
	if err != nil {
		logrus.Errorln("Init repository", err)
		panic(err)
	}
	controller, err := controller.NewController(repository)
	if err != nil {
		logrus.Errorln("Init controller", err)
		panic(err)
	}
	api, err := api.NewApi(controller)
	if err != nil {
		logrus.Errorln("Init api", err)
		panic(err)
	}

	r := gin.Default()
	r.GET("/balance", api.Balance)
	r.POST("/balance", api.Enrollment)
	r.POST("/transfer", api.Transfer)
	r.POST("/order", api.Order)
	r.POST("/order/success", api.OrderSuccess)
	r.POST("/order/failed", api.OrderFailed)
	r.POST("/report", api.Report)
	r.GET("/report/csv", api.CsvReport)
	r.GET("/history", api.History)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	err = r.Run(":8080")
	if err != nil {
		logrus.Errorln("Router run: ", err)
		panic(err)
	}
}
