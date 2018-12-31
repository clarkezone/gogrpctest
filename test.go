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

func (s *HelloServer) SayHello(context.Context, *jamestestrpc.Empty) (*jamestestrpc.TheHello, error) {
	fmt.Println("SayHello")
	return &jamestestrpc.TheHello{Jamesmessage: "Boooooo!"}, nil
}

func servegRPC() {
	lis, err := net.Listen("tcp", ":8282")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	jamestestrpc.RegisterJamesTestServiceServer(grpcServer, &HelloServer{})
	grpcServer.Serve(lis)
}

func startclient() {
	fmt.Println("client")

	conn, err := grpc.Dial("8282", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer conn.Close()

	client := jamestestrpc.NewJamesTestServiceClient(conn)
	result, err := client.SayHello(context.Background(), &jamestestrpc.Empty{})

	if err != nil {
		if result == nil {
			log.Fatal("Result was nil")
		}
		fmt.Println(result.Jamesmessage)
	} else {
		fmt.Println("There was an error")
	}
}

func main() {
	//- [x] Hello run in docker
	fmt.Println("Serve")
	//- [x] go modules
	//- [x] let's encrypt domain
	//- [x] let's encrypt domain dockerImage
	//serveHttps()
	//- [ ] basic gRPC
	//servegRPC()
	startclient()
	//= [ ] basic gRPC with let's encrypt
	//- [ ] gRPC with encryped static auth
	//- [ ] Objective-C client
	//- [ ] gRPC push
	//- [ ] Redis based auth (or another SSO)
	//- [ ] Promethius monitoring
	//- [ ] intellisense in VIM
}
