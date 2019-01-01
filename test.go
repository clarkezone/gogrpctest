package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	yaml "gopkg.in/yaml.v2"

	"github.com/clarkezone/gotest/jamestestrpc"
	"golang.org/x/crypto/acme/autocert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
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

func serveHttps(serverName string, serverPort int) {
	fmt.Println("Serving HTTPS")
	m := &autocert.Manager{
		Cache:      autocert.DirCache("TLS"),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(serverName),
	}
	s := &http.Server{
		Addr:      fmt.Sprintf(":%v", serverPort),
		TLSConfig: m.TLSConfig(),
		Handler:   http.HandlerFunc(myHandler),
	}
	log.Fatal(s.ListenAndServeTLS("", ""))
}

type conf struct {
	ServerPort    int    `yaml:"serverport"`
	TlsServerName string `yaml:"tlsservername"`
	ClientPort    int    `yaml:"clientport"`
}

func (c *conf) getConf() {

	yamlFile, err := ioutil.ReadFile("conf.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
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

func servegRPCAutoCert(serverName string, serverPort int) {
	fmt.Println("Serving gRPC AutoCert")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", serverPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer, err := listenWithAutoCert(serverName, 0)
	if err != nil {
		log.Fatalf("failed to listenwithautocert: %v", err)
	}
	helloServer := HelloServer{}
	jamestestrpc.RegisterJamesTestServiceServer(grpcServer, &helloServer)
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("failed to serve grpc with autocert: %v", err)
	}
}

func listenWithAutoCert(serverName string, p int) (*grpc.Server, error) {
	m := &autocert.Manager{
		Cache:      autocert.DirCache("tls"),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(serverName),
	}
	//todo: do we actually need to listen here to get autocert to work?  If yes, put port in config
	go http.ListenAndServe(":8080", m.HTTPHandler(nil))
	creds := credentials.NewTLS(&tls.Config{GetCertificate: m.GetCertificate})
	srv := grpc.NewServer(grpc.Creds(creds))
	reflection.Register(srv)
	return srv, nil
}

func listenBasic(p int) (net.Listener, error) {
	lis, err := net.Listen("tcp", ":8282")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	return lis, err
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

func startclientsecure(servername string, port int) {
	fmt.Println("Client Secure")

	conf := &tls.Config{ServerName: servername}

	creds := credentials.NewTLS(conf)

	conn, err := grpc.Dial(fmt.Sprintf("%v:%v", servername, port), grpc.WithTransportCredentials(creds))
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
	var c conf
	c.getConf()
	//- [x] Hello run in docker
	//- [x] go modules
	//- [x] let's encrypt domain
	//- [x] let's encrypt domain dockerImage
	//serveHttps()
	//- [x] basic gRPC
	//servegRPC()
	//startclientsecure(c.TlsServerName, c.ClientPort)
	//= [ ] basic gRPC with let's encrypt
	servegRPCAutoCert(c.TlsServerName, c.ServerPort)
	//- [ ] gRPC with encryped static auth and YAML config for UN/PW/secure etc
	//- [ ] Objective-C watch client
	//- [ ] gRPC streaming / push time tick
	//- [ ] Promethius monitoring
	//- [ ] Pluggable / Redis based auth (or another SSO)
}
