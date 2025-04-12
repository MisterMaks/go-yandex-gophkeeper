package main

import (
	"log"

	pb "github.com/MisterMaks/go-yandex-gophkeeper/api/proto/service"
	internal_config "github.com/MisterMaks/go-yandex-gophkeeper/internal/client/config"
	"github.com/MisterMaks/go-yandex-gophkeeper/internal/client/infrastructure/client"
	"github.com/MisterMaks/go-yandex-gophkeeper/internal/client/ui"
	internal_usecase "github.com/MisterMaks/go-yandex-gophkeeper/internal/client/usecase"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const PathToCertificate = "certificates/cert.pem"

var BuildVersion string
var BuildDate string

func printBuildInfo() {
	if BuildVersion == "" {
		log.Println("Build version: N/A")
	} else {
		log.Println("Build version:", BuildVersion)
	}

	if BuildDate == "" {
		log.Println("Build date: N/A")
	} else {
		log.Println("Build date:", BuildDate)
	}
}

func main() {
	printBuildInfo()

	config, err := internal_config.New()
	if err != nil {
		log.Fatalln("CRITICAL\tFailed to create config. Error:", err)
	}

	tlsCert, err := credentials.NewClientTLSFromFile(PathToCertificate, "")
	if err != nil {
		log.Fatalln("Failed to get TLS certificate. Error:", err)
	}

	cc, err := grpc.NewClient(config.GRPCAddress, grpc.WithTransportCredentials(tlsCert))
	if err != nil {
		log.Fatalln(err)
	}
	defer cc.Close()

	c := pb.NewGoYandexGophkeeperClient(cc)
	gophkeeperClient := client.NewGophkeeperClient(c)
	usecase := internal_usecase.NewUsecase(gophkeeperClient)

	tui := ui.NewTUI(usecase)
	err = tui.Run()
	if err != nil {
		log.Fatalln("Failed to run client")
	}
}
