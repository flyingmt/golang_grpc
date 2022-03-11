package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	_ "time"

	"github.com/flyingmt/pcbook/pb"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// maxinum 1 megabyte
const maxImageSize = 1 << 20

// LaptopServer is the server that provides laptop service
type LaptopServer struct {
    laptopStore LaptopStore
    imageStore ImageStore
}

// NewLaptopServer returns a new LaptopServer
func NewLaptopServer(laptopStore LaptopStore, imageStore ImageStore) *LaptopServer {
    return &LaptopServer{laptopStore, imageStore}
}

// CreateLaptop is a unary RPC to create a new laptop
func (server *LaptopServer) CreateLaptop(
    ctx context.Context, 
    req *pb.CreateLaptopRequest,
) (*pb.CreateLaptopResponse, error) {
    laptop := req.GetLaptop()
    log.Printf("received a create-laptop request with id: %s\n", laptop.Id)

    if len(laptop.Id) > 0 {
        // check if it's a valid UUID
        _, err := uuid.Parse(laptop.Id)
        if err != nil {
            return nil, status.Errorf(codes.InvalidArgument, "laptop ID is not a valid UUID: %v", err)
        }
    } else {
        id, err := uuid.NewRandom()
        if err != nil {
            return nil, status.Errorf(codes.Internal, "cannot generate a new laptop ID: %v", err)
        }
        laptop.Id = id.String()
    }

    // some heavy processing
    // time.Sleep(10 * time.Second)

    if err := contextError(ctx); err != nil {
        return nil, err
    }

    // save the laptop to in-memory store (must be db in production)
    err := server.laptopStore.Save(laptop)
    if err != nil {
        code := codes.Internal
        if errors.Is(err, ErrAlreadyExists) {
            code = codes.AlreadyExists
        }

        return nil, status.Errorf(code, "cannot save laptop to the store: %v", err)
    }

    log.Printf("saved laptop with id: %s\n", laptop.Id)

    res := &pb.CreateLaptopResponse{
        Id: laptop.Id,
    }

    return res, nil
}

// SearchLaptop is a server-streaming RPC to search for laptops
func (server *LaptopServer) SearchLaptop(
    req *pb.SearchLaptopRequest,
    stream pb.LaptopService_SearchLaptopServer,
) error {
    filter := req.GetFilter()
    log.Printf("receive a search-laptop request with filter: %v\n", filter)

    err := server.laptopStore.Search(
        stream.Context(),
        filter,
        func(laptop *pb.Laptop) error {
            res := &pb.SearchLaptopResponse{Laptop: laptop}

            err := stream.Send(res)
            if err != nil {
                return err
            }

            log.Printf("sent laptop with id: %s\n", laptop.GetId())
            return nil
        },
    )

    if err != nil {
        return status.Errorf(codes.Internal, "unexpected error: %v", err)
    }

    return nil
}

func (server *LaptopServer) UploadImage(stream pb.LaptopService_UploadImageServer) error {
    req, err := stream.Recv()
    if err != nil {
        return logError(status.Errorf(codes.Unknown, "cannot receive image info"))
    }

    laptopID := req.GetInfo().GetLaptopId()
    imageType := req.GetInfo().GetImageType()
    log.Printf("receive an upload-image request for laptop %s with image type %s", laptopID, imageType)

    laptop, err := server.laptopStore.Find(laptopID)
    if err != nil {
        return logError(status.Errorf(codes.Internal, "cannot find laptop: %v", err))
    }
    if laptop == nil {
        return logError(status.Errorf(codes.InvalidArgument, "laptop %s doesn't exist", laptopID))
    }

    imageData := bytes.Buffer{}
    imageSize := 0

    for {
        if err := contextError(stream.Context()); err != nil {
            return err
        }


        log.Print("waiting to receive more data")

        req, err := stream.Recv()
        if err == io.EOF {
            log.Print("no more data")
            break
        }
        if err != nil {
            return logError(status.Errorf(codes.Unknown, "cannot receive chunk data: %v", err))
        }

        chunck := req.GetChunkData()
        size := len(chunck)

        log.Printf("received a chunk with size: %d", size)

        imageSize += size
        if imageSize > maxImageSize {
            return logError(status.Errorf(codes.InvalidArgument, "image is too large: %d > %d", imageSize, maxImageSize))
        }

        // write slowly
        // time.Sleep(time.Second)

        _, err = imageData.Write(chunck)
        if err != nil {
            return logError(status.Errorf(codes.Internal, "cannot write chuck data: %v", err))
        }
    }

    imageID, err := server.imageStore.Save(laptopID, imageType, imageData)
    if err != nil {
        return logError(status.Errorf(codes.Internal, "cannot save image to the store: %v", err))
    }

    res := &pb.UploadImageResponse{
        Id: imageID,
        Size: uint32(imageSize),
    }

    err = stream.SendAndClose(res)
    if err != nil {
        return logError(status.Errorf(codes.Unknown, "cannot send response: %v", err))
    }

    log.Printf("saved image with id: %s, size: %d\n", imageID, imageSize)
    return nil
}

func contextError(ctx context.Context) error {
    switch ctx.Err() {
    case context.Canceled:
        return logError(status.Error(codes.Canceled, "request is canceled"))
    case context.DeadlineExceeded:
        return logError(status.Error(codes.DeadlineExceeded, "deadline is exceeded"))
    default:
        return nil
    }
}


func logError(err error) error {
    if err != nil {
        log.Print(err)
    }
    return err
}
