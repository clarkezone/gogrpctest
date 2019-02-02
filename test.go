package test

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"

	"github.com/clarkezone/gotest/jamestestrpc"
	"golang.org/x/crypto/acme/autocert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
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
	Secret        string `yaml:"secret"`
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

func (s *HelloServer) SayHelloStreaming(stream jamestestrpc.JamesTestService_SayHelloStreamingServer) error {
	fmt.Println("Server:SayHelloStreaming")
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("Server:hit EOF, terminating")
			return nil
		}
		if err != nil {
			fmt.Printf("Server:hit error, terminating this call: %v %v", err)
			return err
		}
		fmt.Printf("Server:Message received: %v\n", in.Jamesmessage)
		if err := stream.Send(&jamestestrpc.TheHello{Jamesmessage: "Boooooo! from server"}); err != nil {
			log.Fatalf("Server: send error: %v", err)
		}
	}
	fmt.Println("Returning nil")
	return nil
}

type Backend struct {
	config *conf
}

func CreateBackend() *Backend {
	bs := &Backend{}
	bs.config = &conf{}
	bs.config.getConf()
	return bs
}

func (be *Backend) ServegRPC() {
	be.servegRPC(be.config.TlsServerName, be.config.ServerPort)
}

func (be *Backend) ServegRPCAutoCert() {
	be.servegRPCAutoCert(be.config.TlsServerName, be.config.ServerPort)
}

func (be *Backend) servegRPC(serverName string, serverPort int) {
	fmt.Printf("Serving gRPC for endpoint %v on port %v\n", serverName, serverPort)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", serverPort))
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

func (be *Backend) servegRPCAutoCert(serverName string, serverPort int) {
	fmt.Printf("Serving gRPC AutoCert for endpoint %v on port %v\n", serverName, serverPort)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", serverPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer, err := be.listenWithAutoCert(serverName, 0)
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

func (be *Backend) listenWithAutoCert(serverName string, p int) (*grpc.Server, error) {
	m := &autocert.Manager{
		Cache:      autocert.DirCache("tls"),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(serverName),
	}
	//todo: do we actually need to listen here to get autocert to work?  If yes, put port in config
	go http.ListenAndServe(":8080", m.HTTPHandler(nil))
	creds := credentials.NewTLS(&tls.Config{GetCertificate: m.GetCertificate})

	opts := []grpc.ServerOption{grpc.Creds(creds),
		grpc.UnaryInterceptor(be.unaryInterceptor),
		grpc.StreamInterceptor(be.streamInterceptor),
	}

	srv := grpc.NewServer(opts...)
	reflection.Register(srv)
	return srv, nil
}

func (be *Backend) unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if info.FullMethod == "/proto.EventStoreService/GetJWT" { //skip auth when requesting JWT

		return handler(ctx, req)
	}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		clientLogin := strings.Join(md["login"], "")

		if clientLogin != be.config.Secret {
			return nil, fmt.Errorf("bad creds")
		}

		//ctx = context.WithValue(ctx, clientIDKey, clientID)
		return handler(ctx, req)
	}

	return nil, fmt.Errorf("missing credentials")
}

func (be *Backend) streamInterceptor(req interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if md, ok := metadata.FromIncomingContext(ss.Context()); ok {
		clientLogin := strings.Join(md["login"], "")

		if clientLogin != be.config.Secret {
			return fmt.Errorf("bad creds")
		}

		return nil
	}
	return fmt.Errorf("missing credentials")
}

func (be *Backend) StartclientStreaming() {
	startclientStreaming(be.config.TlsServerName, be.config.ClientPort)
}

func (be *Backend) StartclientSecure() {
	startclientsecure(be.config.TlsServerName, be.config.ClientPort, be.config.Secret)
}

func (be *Backend) StartclientStreamingSecure() {
	startclientStreamingSecure(be.config.TlsServerName, be.config.ClientPort, be.config.Secret)
}

func listenBasic(p int) (net.Listener, error) {
	lis, err := net.Listen("tcp", ":8282")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	return lis, err
}

func Startclient(servername string, port int) {
	fmt.Println("Client")

	conn, err := grpc.Dial(fmt.Sprintf("%v:%v", servername, port), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer conn.Close()

	client := jamestestrpc.NewJamesTestServiceClient(conn)

	result, err := client.SayHello(context.Background(), &jamestestrpc.TheHello{})

	if err != nil {
		log.Fatalf("Error calling RPC: %v", err)
	}

	if result == nil {
		log.Fatal("No error but Result was nil")
	}

	fmt.Println(result.Jamesmessage)
}

func startclientStreaming(servername string, port int) {
	fmt.Printf("Client Streaming %v %v\n", servername, port)

	conn, err := grpc.Dial(fmt.Sprintf("%v:%v", servername, port), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Client Error: %v", err)
	}
	defer conn.Close()

	client := jamestestrpc.NewJamesTestServiceClient(conn)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	//defer cancel()
	stream, err := client.SayHelloStreaming(ctx)
	if err != nil {
		log.Fatalf("Client Error calling RPC: %v", err)
	}
	waitc := make(chan struct{})
	go func() {
		in, err := stream.Recv()
		if err != nil {
			log.Fatalf("Client Error: %v", err)
		}
		if in == nil {
			fmt.Println("Nothing received")
		} else {
			fmt.Printf("Client Received: %v\n", in.Jamesmessage)
		}
		close(waitc)
	}()

	err = stream.Send(&jamestestrpc.TheHello{Jamesmessage: "FromFlicnet!"})
	if err != nil {
		log.Fatalf("Send error %v\n", err)
	}
	fmt.Println("Client: message sent")
	<-waitc
	fmt.Println("Client: wait done, closing send")
	stream.CloseSend()
}

func startclientStreamingSecure(servername string, port int, keyword string) {
	fmt.Printf("Client Streaming %v %v\n", servername, port)

	conf := &tls.Config{ServerName: servername}

	creds := credentials.NewTLS(conf)

	auth := Authentication{Login: keyword}

	conn, err := grpc.Dial(fmt.Sprintf("%v:%v", servername, port), grpc.WithTransportCredentials(creds),
		grpc.WithPerRPCCredentials(&auth))
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer conn.Close()

	client := jamestestrpc.NewJamesTestServiceClient(conn)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	//defer cancel()
	stream, err := client.SayHelloStreaming(ctx)
	if err != nil {
		log.Fatalf("Client Error calling RPC: %v", err)
	}
	waitc := make(chan struct{})
	go func() {
		in, err := stream.Recv()
		if err != nil {
			log.Fatalf("Client Error: %v", err)
		}
		if in == nil {
			fmt.Println("Nothing received")
		} else {
			fmt.Printf("Client Received: %v\n", in.Jamesmessage)
		}
		close(waitc)
	}()

	err = stream.Send(&jamestestrpc.TheHello{Jamesmessage: "FromFlicnet!"})
	if err != nil {
		log.Fatalf("Send error %v\n", err)
	}
	fmt.Println("Client: message sent")
	<-waitc
	fmt.Println("Client: wait done, closing send")
	stream.CloseSend()
}

type Authentication struct {
	Login string
}

func (a *Authentication) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{
		"login": a.Login,
	}, nil
}

func (a *Authentication) RequireTransportSecurity() bool {
	return true
}

func startclientsecure(servername string, port int, keyword string) {
	fmt.Printf("Client Secure: servername:%v port:%v keyword:%v\n", servername, port, keyword)

	conf := &tls.Config{ServerName: servername}

	creds := credentials.NewTLS(conf)

	auth := Authentication{Login: keyword}

	conn, err := grpc.Dial(fmt.Sprintf("%v:%v", servername, port), grpc.WithTransportCredentials(creds),
		grpc.WithPerRPCCredentials(&auth))
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer conn.Close()

	client := jamestestrpc.NewJamesTestServiceClient(conn)
	result, err := client.SayHello(context.Background(), &jamestestrpc.TheHello{})

	if err != nil {
		log.Fatalf("Error calling RPC: %v", err)
	}

	if result == nil {
		log.Fatal("No error but Result was nil")
	}

	fmt.Println(result.Jamesmessage)
}

//func main() {
//var c conf
//c.getConf()
//- [x] Hello run in docker
//- [x] go modules
//- [x] let's encrypt domain
//- [x] let's encrypt domain dockerImage
//serveHttps()
//- [x] basic gRPC
//servegRPC()
//startclientsecure(c.TlsServerName, c.ClientPort)
//= [x] basic gRPC with let's encrypt
//servegRPCAutoCert(c.TlsServerName, c.ServerPort)
//- [ ] gRPC with encryped static auth and YAML config for UN/PW/secure etc
//- [ ] gRPC streaming / push time tick
//- [ ] Promethius monitoring
//- [ ] Objective-C client
//- [ ] Dart client
//- [ ] Pluggable / Redis based auth (or another SSO)
//}
