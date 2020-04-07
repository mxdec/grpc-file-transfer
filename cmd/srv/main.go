package main

import (
	"github.com/sirupsen/logrus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/mxdec/grpc-file-transfer/srv"
)

var (
	// log
	logLevel = kingpin.Flag("log.level", "Log level.").Default("info").Enum("trace", "debug", "info", "error")

	// tls
	tlsPathCert = kingpin.Flag("tls.path-cert", "Path to signed certificate file.").Default("").String()
	tlsPathKey  = kingpin.Flag("tls.path-key", "Path to certificate key file.").Default("").String()

	// net
	httpAddr = kingpin.Flag("http.listen-address", "HTTP address to listen.").Default("127.0.0.1:8080").String()
	grpcAddr = kingpin.Flag("grpc.listen-address", "gRPC address to listen.").Default("127.0.0.1:8081").String()
)

func init() {
	kingpin.Parse()
	lvl, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		logrus.Panic(err)
	}
	logrus.SetLevel(lvl)
}

func main() {
	kingpin.Version("1.0.0")
	logrus.Info("Starting gRPC_Demo")
	srv := srv.New(
		*httpAddr,
		*grpcAddr,
		*tlsPathCert,
		*tlsPathKey,
	)
	srv.Run()
}
