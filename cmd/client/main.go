package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/flyingmt/pcbook/client"
	"github.com/flyingmt/pcbook/pb"
	"github.com/flyingmt/pcbook/sample"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func testCreateLaptop(laptopClient *client.LaptopClient) {
	laptopClient.CreateLaptop(sample.NewLaptop())
}

func testSearchLaptop(laptopClient *client.LaptopClient) {
	for i := 0; i < 10; i++ {
		laptopClient.CreateLaptop(sample.NewLaptop())
	}

	filter := &pb.Filter{
		MaxPriceUsd: 3000,
		MinCpuCores: 4,
		MinCpuGhz:   2.5,
		MinRam:      &pb.Memory{Value: 8, Unit: pb.Memory_GIGABYTE},
	}

	laptopClient.SearchLaptop(filter)

}

func testUploadImage(laptopClient *client.LaptopClient) {
	laptop := sample.NewLaptop()
	laptopClient.CreateLaptop(laptop)
	laptopClient.UploadImage(laptop.GetId(), "image/laptop.jpg")
}

func testRateLaptop(laptopClient *client.LaptopClient) {
    n := 3
    laptopIDs := make([]string, n)

    for i := 0; i < n; i++ {
        laptop := sample.NewLaptop()
        laptopIDs[i] = laptop.GetId()
        laptopClient.CreateLaptop(laptop)
    }

    scores := make([]float64, n)
    for {
        fmt.Print("Rate Laptop (y/n)? ")
        var answer string
        fmt.Scan(&answer)

        if strings.ToLower(answer) != "y" {
            break
        }

        for i := 0; i < n; i++ {
            scores[i] = sample.RandomLaptopScore()
        }

        err := laptopClient.RateLaptop(laptopIDs, scores)
        if err != nil {
            log.Fatal(err)
        }
    }
}

const (
	username = "admin1"
	//username = "user1"
	password = "secret"
	refreshDuration = 30 * time.Second
)

func authMethods() map[string]bool {
	const laptopServicePath = "/LaptopService/"
	
	return map[string]bool {
		laptopServicePath + "CreateLaptop": true,
		laptopServicePath + "UploadImage": true,
		laptopServicePath + "RateLaptop": true,
	}
}

func main() {
	fmt.Println("Hello world from Client")
	serverAddress := flag.String("address", "", "the server address")
	flag.Parse()
	log.Printf("dial server %s\n", *serverAddress)

	cc1, err := grpc.Dial(*serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("cannot dial server: ", err)
	}

	authClient := client.NewAuthClient(cc1, username, password)
	interceptor, err := client.NewAuthInterceptor(authClient, authMethods(), refreshDuration)
	if err != nil {
		log.Fatal("Cannot create auth interceptor: ", err)
	}

	cc2, err := grpc.Dial(
		*serverAddress, 
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(interceptor.Unary()),
		grpc.WithStreamInterceptor(interceptor.Stream()),
	)
	if err != nil {
		log.Fatal("cannot dial server: ", err)
	}

	laptopClient := client.NewLaptopClient(cc2)

	//testCreateLaptop(laptopClient)
	//testSearchLaptop(laptopClient)
	//testUploadImage(laptopClient)
    testRateLaptop(laptopClient)
}


