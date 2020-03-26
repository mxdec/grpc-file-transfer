package srv

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pb "github.com/maxdcr/grpc-file-transfer/proto"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// Server ...
type Server struct {
	// Address
	httpAddress string
	grpcAddress string

	// HTTP
	h *http.Server

	// gRPC
	grpcSrv *grpc.Server

	// TCP
	lst net.Listener
	tls TLSConfig
}

// TLSConfig provide certificates path for gRPC server
type TLSConfig struct {
	PathCert string
	PathKey  string
}

// New returns new server
func New(
	httpAddress string,
	grpcAddress string,
	tlsPathCert string,
	tlsPathKey string,
) *Server {
	return &Server{
		httpAddress: httpAddress,
		grpcAddress: grpcAddress,
		tls: TLSConfig{
			PathCert: tlsPathCert,
			PathKey:  tlsPathKey,
		},
	}
}

func (srv *Server) health(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte(`{"status": "ok"}`))
}

// Run server server
func (srv *Server) startHTTP() {
	router := mux.NewRouter()
	router.HandleFunc("/health", srv.health).Methods(http.MethodGet)

	srv.h = &http.Server{
		Addr:    srv.httpAddress,
		Handler: router,
	}
	go func() {
		if err := srv.h.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("http: %s\n", err)
		}
	}()
	logrus.Infof("listen HTTP: %s", srv.httpAddress)
}

func (srv *Server) startGRPC() error {
	var grpcOpts []grpc.ServerOption
	var err error

	// open TCP listener
	srv.lst, err = net.Listen("tcp", srv.grpcAddress)
	if err != nil {
		logrus.Fatalf("failed to listen: %v", err)
	}

	if len(srv.tls.PathCert) > 0 {
		creds, err := credentials.NewServerTLSFromFile(srv.tls.PathCert, srv.tls.PathKey)
		if err != nil {
			return err
		}
		grpcOpts = append(grpcOpts, grpc.Creds(creds))
	}
	srv.grpcSrv = grpc.NewServer(grpcOpts...)
	pb.RegisterApiServer(srv.grpcSrv, &APIServer{})
	grpc_health_v1.RegisterHealthServer(srv.grpcSrv, &HealthServer{Service: "gRPCDemo"})

	// start gRPC server
	go func() {
		if err := srv.grpcSrv.Serve(srv.lst); err != nil {
			logrus.Fatalf("grpc: %s\n", err)
		}
	}()
	logrus.Infof("listen gRPC: %s", srv.grpcAddress)
	return nil
}

// shutdown closes HTTP and gRPC servers
func (srv *Server) shutdown() error {
	// close HTTP
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// extra handling here
		cancel()
	}()

	if err := srv.h.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shudown http server: %v", err)
	}
	srv.grpcSrv.Stop()
	if err := srv.lst.Close(); err != nil {
		return fmt.Errorf("failed to close TCP listener: %v", err)
	}
	return nil
}

// Run server
func (srv *Server) Run() error {
	srv.startHTTP()
	srv.startGRPC()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(
		signalChan,
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	// wait signal to shutdown server
	<-signalChan
	logrus.Info("stopping server...")
	return srv.shutdown()
}
