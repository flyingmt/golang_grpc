package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/flyingmt/pcbook/pb"
	"github.com/flyingmt/pcbook/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)




const (
	secretKey = "secret"
	tokenDuration = 15 * time.Minute
)

func seedUsers(userStore service.UserStore) error {
	err := createUser(userStore, "admin1", "secret", "admin")
	if err != nil {
		return err
	}

	return createUser(userStore, "user1", "secret", "user")
}

func createUser(userStore service.UserStore, username string, password string, role string) error {
	user, err := service.NewUser(username, password, role)
	if err != nil {
		return err
	}

	return userStore.Save(user)
}

func accessibleRoles() map[string][]string {
	const laptopServicePath = "/LaptopService/"
	
	return map[string][]string {
		laptopServicePath + "CreateLaptop": {"admin"},
		laptopServicePath + "UploadImage": {"admin"},
		laptopServicePath + "RateLaptop": {"admin", "user"},
	}
}

func main() {
	fmt.Println("Hello World from Server")
	port := flag.Int("port", 0, "the server port")
	flag.Parse()
	log.Printf("start server on port %d\n", *port)

	userStore := service.NewInMemoryUserStore()
	err := seedUsers(userStore)
	if err != nil {
		log.Fatal("Cannot seed users")
	}

	jwtManager := service.NewJWTManager(secretKey, tokenDuration)
	authServer := service.NewAuthServer(userStore, jwtManager)

	laptopStore := service.NewInMemoryLaptopStore()
	imageStore := service.NewDiskImageStore("img")
    ratingStore := service.NewInMemoryRatingStore()
	laptopServer := service.NewLaptopServer(laptopStore, imageStore, ratingStore)

	interceptor := service.NewAuthInterceptor(jwtManager, accessibleRoles())
	grpcServer := grpc.NewServer(
        grpc.UnaryInterceptor(interceptor.Unary()),
        grpc.StreamInterceptor(interceptor.Stream()),
    )

	pb.RegisterAuthServiceServer(grpcServer, authServer)
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)
    reflection.Register(grpcServer)

	address := fmt.Sprintf("0.0.0.0:%d", *port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("cannot start server: ", err)
	}

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot start server: ", err)
	}
}
