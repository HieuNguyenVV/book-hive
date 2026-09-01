package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/HieuNguyenVV/book-hive/docs"
	"github.com/HieuNguyenVV/book-hive/internal/pkg/config"
	"github.com/HieuNguyenVV/book-hive/internal/server"
)

// @title           Book Hive API
// @version         1.0
// @description     Book Hive REST API
// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey AccessToken
// @in              header
// @name            Authorization
// @securityDefinitions.apikey RefreshToken
// @in              header
// @name            RefreshToken
func main() {
	config, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		cancel()
	}()

	clientContainer := server.NewClientContainer(config)
	repositoryContainer := server.NewRepositoryContainer(clientContainer)
	serviceContainer := server.NewServiceContainer(clientContainer, repositoryContainer)
	controllerContainer := server.NewControllerContainer(clientContainer, serviceContainer)
	srv := server.NewServer(clientContainer, repositoryContainer, serviceContainer, controllerContainer)
	srv.Run(ctx)
}
