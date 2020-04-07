package srv

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/mxdec/grpc-file-transfer/proto"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// HealthServer implements grpc_health_v1.HealthServer
type HealthServer struct {
	Service string
}

// Check ...
func (hs *HealthServer) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	if req.Service == "" || req.Service == hs.Service {
		return &grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_SERVING}, nil
	}
	return &grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_NOT_SERVING}, nil
}

// Watch ...
func (hs *HealthServer) Watch(*grpc_health_v1.HealthCheckRequest, grpc_health_v1.Health_WatchServer) error {
	return nil
}

// APIServer ...
type APIServer struct{}

// GetFile ...
func (a *APIServer) GetFile(c context.Context, r *pb.GetFileRequest) (*pb.File, error) {
	fmt.Println(r.GetFilePath())
	return &pb.File{
		Name:        "lorem",
		ContentType: "plain/text",
		Content:     []byte("implement me"),
	}, nil
}

// SetFile ...
func (a *APIServer) SetFile(c context.Context, r *pb.SetFileRequest) (*pb.File, error) {
	if file := r.GetFileContent(); file != nil {
		fmt.Println(file.GetName(), file.GetContentType())
		fmt.Println(string(file.GetContent()))
		return file, nil
	}
	return nil, errors.New("empty file")
}
