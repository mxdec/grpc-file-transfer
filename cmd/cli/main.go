package main

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	pb "github.com/maxdcr/grpc-file-transfer/proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	// api
	apiAddress = kingpin.Flag("api.address", "Scaleway API Address").Default("127.0.0.1:8080").String()

	// file push
	file         = kingpin.Command("file", "Manage files.")
	filePush     = file.Command("push", "Push a local file to destination.")
	filePushSrc  = filePush.Arg("src", "Source location.").Required().String()
	filePushPath = filePush.Arg("path", "Destination location.").Required().String()

	// file get
	fileCat     = file.Command("cat", "Display file from remote location.")
	fileCatPath = fileCat.Arg("path", "Remote file file path.").Required().String()

	// regex
	re = regexp.MustCompile(`^([a-z0-9_-]{1,64}):([a-z0-9_/-]{1,256}).(crt|key|yml|json)$`)

	// gRPC client
	api pb.ApiClient
	ctx context.Context
)

func init() {
	kingpin.Parse()

	// creds := credentials.NewTLS(&tls.Config{})
	conn, err := grpc.Dial(*apiAddress, grpc.WithInsecure())
	if err != nil {
		logrus.Fatalf("Missing required context to run tests: %v", err)
	}
	ctx = context.Background()

	api = pb.NewApiClient(conn)
}

func main() {
	kingpin.Version("1.0.0")
	switch kingpin.Parse() {
	case filePush.FullCommand():
		handleFilePush(*filePushSrc, *filePushPath)
	case fileCat.FullCommand():
		handleFileCat(*fileCatPath)
	default:
		kingpin.Usage()
	}
}

type specsDest struct {
	namespace   string
	path        string
	contentType string
}

var handlerMap = map[*regexp.Regexp]func(string) (string, error){
	regexp.MustCompile(`^([a-z0-9_-]{1,64}).(yml)$`):     parseYaml,
	regexp.MustCompile(`^([a-z0-9_-]{1,64}).(json)$`):    parseJson,
	regexp.MustCompile(`^([a-z0-9_-]{1,16}).(crt|key)$`): parseCerts,
}

// content types
const (
	typeYaml = "application/yaml"
	typeJson = "application/json"
	typeText = "plain/text"
)

func parseYaml(src string) (string, error) {
	fmt.Println("this is a YAML file")
	return typeYaml, nil
}

func parseJson(src string) (string, error) {
	fmt.Println("this is a JSON file")
	return typeJson, nil
}

func parseCerts(src string) (string, error) {
	file, err := ioutil.ReadFile(src)
	if err != nil {
		return "", err
	}
	block, _ := pem.Decode(file)
	if block == nil {
		return "", fmt.Errorf("%s failed to decode PEM", src)
	}
	switch {
	case strings.HasSuffix(src, ".crt"):
		_, err = x509.ParseCertificate(block.Bytes)
	case strings.HasSuffix(src, ".key"):
		_, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	default:
		err = fmt.Errorf("%s is not a cert", src)
	}
	return typeText, err
}

func lookupParser(src, dest string) (string, error) {
	for regex, handler := range handlerMap {
		if regex.MatchString(dest) {
			return handler(src)
		}
	}
	return "", errors.New("invalid file type")
}

func handleFilePush(src string, dest string) {
	specs, err := parseSpecs(dest)
	if err != nil {
		logrus.Fatal(err)
	}

	specs.contentType, err = lookupParser(src, specs.path)
	if err != nil {
		logrus.Fatal(err)
	}

	if err := pushFile(src, specs); err != nil {
		logrus.Fatal(err)
	}

	logrus.Printf("%s has been pushed to %s\n", src, dest)
}

func pushFile(src string, specs specsDest) error {
	file, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	_, err = api.SetFile(ctx, &pb.SetFileRequest{
		Namespace: specs.namespace,
		FilePath:  specs.path,
		FileContent: &pb.File{
			Name:        filepath.Base(specs.path),
			ContentType: specs.contentType,
			Content:     file,
		},
	})
	return err
}

func parseSpecs(dest string) (specsDest, error) {
	var specs specsDest

	if re.MatchString(dest) == true {
		path := strings.Split(dest, ":")
		specs.namespace = path[0]
		specs.path = path[1]
		return specs, nil
	}
	return specs, errors.New("invalid arguments")
}

func handleFileCat(dest string) {
	logrus.Print("Implement me!")
}
