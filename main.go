package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ramailh/authentication-server/delivery/http/controller"
	"github.com/ramailh/authentication-server/repository/mongodb"
	"github.com/ramailh/authentication-server/service"

	"github.com/gin-gonic/gin"
	"github.com/subosito/gotenv"
)

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	env := ".env"
	if len(os.Args) > 1 {
		env = os.Args[1]
	}

	if err := gotenv.Load(env); err != nil {
		log.Println(err)
	}
}

func main() {
	errs := make(chan error)

	client, err := mongodb.ConnectMongo(os.Getenv("MONGO_HOST"), os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD"))
	if err != nil {
		log.Fatalln(err)
	}

	repo := mongodb.NewMongoRepo(client)
	srv := service.NewUserService(repo)

	go func() {
		c := make(chan os.Signal, 1)

		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

		errs <- fmt.Errorf("%v", <-c)
	}()

	go func() {

		rtr := gin.Default()

		controller.NewUserRoute(rtr, srv)

		if err = rtr.Run(":" + os.Getenv("PORT")); err != nil {
			errs <- err
		}
	}()

	log.Fatal(<-errs)
}
