package main

import (
	"log"

	pb "github.com/MisterMaks/go-yandex-gophkeeper/api/proto/service"
	"github.com/MisterMaks/go-yandex-gophkeeper/internal/client/config"
	"github.com/MisterMaks/go-yandex-gophkeeper/internal/client/infrastructure/client"
	"github.com/MisterMaks/go-yandex-gophkeeper/internal/client/ui"
	"github.com/MisterMaks/go-yandex-gophkeeper/internal/client/usecase"
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

	c, err := config.New()
	if err != nil {
		log.Fatalln("CRITICAL\tFailed to create config. Error:", err)
	}

	tlsCert, err := credentials.NewClientTLSFromFile(PathToCertificate, "")
	if err != nil {
		log.Fatalln("Failed to get TLS certificate. Error:", err)
	}

	grpcClient, err := grpc.NewClient(c.GRPCAddress, grpc.WithTransportCredentials(tlsCert))
	if err != nil {
		log.Fatalln(err)
	}
	defer grpcClient.Close()

	grpcServiceClient := pb.NewGoYandexGophkeeperClient(grpcClient)
	gophkeeperClient := client.NewGophkeeperClient(grpcServiceClient)
	u := usecase.NewUsecase(gophkeeperClient)

	tui := ui.NewTUI(u)
	err = tui.Run()
	if err != nil {
		log.Fatalln("Failed to run client")
	}
}
