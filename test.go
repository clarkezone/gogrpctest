package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/clarkezone/gotest/jamestestrpc"
	"golang.org/x/crypto/acme/autocert"
	"google.golang.org/grpc"
)

func myHandler(w http.ResponseWriter, req *http.Request) {
	t := time.Now()
	io.WriteString(w, t.String())
	io.WriteString(w, " hello!\n")
}
func serveHttp() {
	fmt.Println("Serving HTTP")
	s := &http.Server{
		Addr:    ":8282",
		Handler: http.HandlerFunc(myHandler),
	}
	log.Fatal(s.ListenAndServe())
}

func serveHttps() {
	fmt.Println("Serving HTTPS")
	m := &autocert.Manager{
		Cache:      autocert.DirCache("secret-dir"),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist("vul3.objectivepixel.io"),
	}
	s := &http.Server{
		Addr:      ":8443",
		TLSConfig: m.TLSConfig(),
		Handler:   http.HandlerFunc(myHandler),
	}
	log.Fatal(s.ListenAndServeTLS("", ""))
}

type HelloServer struct {
}

func (s *HelloServer) SayHello(context.Context, *jamestestrpc.TheHello) (*jamestestrpc.TheHello, error) {
	fmt.Println("SayHello")
	return &jamestestrpc.TheHello{Jamesmessage: "Boooooo!"}, nil
}

func servegRPC() {
	fmt.Println("Serving gRPC")
	lis, err := net.Listen("tcp", ":8282")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	helloServer := HelloServer{}
	jamestestrpc.RegisterJamesTestServiceServer(grpcServer, &helloServer)
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func startclient() {
	fmt.Println("Client")

	conn, err := grpc.Dial(":8282", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer conn.Close()

	client := jamestestrpc.NewJamesTestServiceClient(conn)
	result, err := client.SayHello(context.Background(), &jamestestrpc.TheHello{})

	if err != nil {
		log.Fatal("Error calling RPC: %v", err)
	}

	if result == nil {
		log.Fatal("No error but Result was nil")
	}

	fmt.Println(result.Jamesmessage)
}

func main() {
	//- [x] Hello run in docker
	//- [x] go modules
	//- [x] let's encrypt domain
	//- [x] let's encrypt domain dockerImage
	//serveHttps()
	//- [x] basic gRPC
	//servegRPC()
	startclient()
	//= [ ] basic gRPC with let's encrypt
	//- [ ] gRPC with encryped static auth and YAML config for UN/PW/secure etc
	//- [ ] Objective-C watch client
	//- [ ] gRPC streaming / push time tick
	//- [ ] Promethius monitoring
	//- [ ] Pluggable / Redis based auth (or another SSO)
}
