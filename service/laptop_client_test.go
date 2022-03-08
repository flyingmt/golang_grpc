package service_test

import (
	"context"
	"net"
	"testing"

	"github.com/flyingmt/pcbook/pb"
	"github.com/flyingmt/pcbook/sample"
	"github.com/flyingmt/pcbook/serializer"
	"github.com/flyingmt/pcbook/service"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestClientCreateLaptop(t *testing.T) {
    t.Parallel()

    laptopServer, serverAddress := startTestLaptopServer(t)
    laptopClient := newTestLaptopClient(t, serverAddress)

    laptop := sample.NewLaptop()
    expectedID := laptop.Id
    req := &pb.CreateLaptopRequest{
        Laptop: laptop,
    }

    res, err := laptopClient.CreateLaptop(context.Background(), req)
    require.NoError(t, err)
    require.NotNil(t, res)
    require.Equal(t, expectedID, res.Id)

    // check that the laptop is saved to the store
    other, err := laptopServer.Store.Find(res.Id)
    require.NoError(t, err)
    require.NotNil(t, other)

    // check that the saved laptop is the same as the one we send
    requireSameLaptop(t, laptop, other)
}

func startTestLaptopServer(t *testing.T) (*service.LaptopServer, string) {
    laptopServer := service.NewLaptopServer(service.NewInMemoryLaptopStore())

    grpcServer := grpc.NewServer()
    pb.RegisterLaptopServiceServer(grpcServer, laptopServer)

    listener, err := net.Listen("tcp", ":0") // random available port
    require.NoError(t, err)

    go grpcServer.Serve(listener) 

    return laptopServer, listener.Addr().String()
}

func newTestLaptopClient(t *testing.T, serverAddress string) pb.LaptopServiceClient {
    conn, err := grpc.Dial(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
    require.NoError(t, err)

    return pb.NewLaptopServiceClient(conn)
}

func requireSameLaptop(t *testing.T, laptop1 *pb.Laptop, laptop2 *pb.Laptop) {
    json1, err := serializer.ProtobufToJSON(laptop1)
    require.NoError(t, err)

    json2, err := serializer.ProtobufToJSON(laptop2)
    require.NoError(t, err)

    require.Equal(t, json1, json2)
}