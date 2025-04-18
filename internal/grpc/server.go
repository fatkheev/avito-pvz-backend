package grpc

import (
    "context"
    "log"
    "net"

    "avito-pvz-service/internal/repository"
    pvz_v1 "avito-pvz-service/internal/grpc/pvz/v1"

    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
    "google.golang.org/protobuf/types/known/timestamppb"
)

type server struct {
    pvz_v1.UnimplementedPVZServiceServer
}

func (s *server) GetPVZList(ctx context.Context, _ *pvz_v1.GetPVZListRequest) (*pvz_v1.GetPVZListResponse, error) {
    pvzs, err := repository.GetAllPVZ()
    if err != nil {
        return nil, err
    }
    resp := &pvz_v1.GetPVZListResponse{}
    for _, p := range pvzs {
        resp.Pvzs = append(resp.Pvzs, &pvz_v1.PVZ{
            Id:               p.ID,
            RegistrationDate: timestamppb.New(p.RegistrationDate),
            City:             p.City,
        })
    }
    return resp, nil
}

func RunGRPCServer() {
    lis, err := net.Listen("tcp", ":3000")
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }
    s := grpc.NewServer()

    pvz_v1.RegisterPVZServiceServer(s, &server{})

    // Включаем reflection, чтобы grpcurl и другие инструменты
    // могли автоматически узнать о сервисах и методах
    reflection.Register(s)

    log.Println("gRPC server is running on port 3000")
    if err := s.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}
