package main

import (
	"net"
	"log"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc"
	"context"
	"net/http"
	"io"
	"medium/proto"
	"io/ioutil"
	"google.golang.org/grpc/connectivity"
	"time"
)

type grpcserv struct{}

func (server *grpcserv)  PostTransaction(context context.Context, model *proto.Request) (*proto.Request, error) {
	log.Print("Handling grpc request")
	return model, nil
}

func main(){
	//s := &grpcserv{}
	//go s.server()

	dial()

	http.HandleFunc("/health", HandleHealth)
	http.ListenAndServe(":8080", nil)
}

func dial() {

	conn, err := grpc.Dial("application.yourdomain:80", grpc.WithInsecure())
	if err != nil {
		log.Print("error: %v", err)
	}
	defer conn.Close()

	i := 0
	for {
		if i > 60 || conn.GetState() == connectivity.Ready {
			break
		}
		log.Print("conn state %s", conn.GetState())
		time.Sleep(time.Millisecond * 100)
		i++
	}

	client := proto.NewTransactionClient(conn)
	resp, err := client.PostTransaction(context.Background(), &proto.Request{Message: "ae"})
	if err != nil {
		log.Print("error while dialing %v", err)
	}
	log.Print(resp)
}

func (server *grpcserv)  server() {
	listenDistributor, err := net.Listen("tcp", ":8333")
	if err != nil {
		log.Print("erro whiel listeting %s", err)
	}
	grpcServer := grpc.NewServer()

	proto.RegisterTransactionServer(grpcServer, server)
	reflection.Register(grpcServer)
	grpcServer.Serve(listenDistributor)
}

var istioHeaders = []string{"x-request-id", "x-b3-traceid", "x-b3-spanid", "x-b3-parentspanid", "x-b3-sampled", "x-b3-flags", "x-ot-span-context"}

func HandleHealth(w http.ResponseWriter, r *http.Request) {

	log.Print("Handling http request")
	if r.Header.Get("propagate") == "" {
		io.WriteString(w, "Ok\n")
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://applicationtwo.istio-apps:8080/health", nil)

	for _, header := range istioHeaders {
		req.Header.Add(header, r.Header.Get(header))
	}
	rs, err := client.Do(req)
	// Process response
	if err != nil {
		panic(err) // More idiomatic way would be to print the error and die unless it's a serious error
	}
	defer rs.Body.Close()

	bodyBytes, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		panic(err)
	}

	bodyString := string(bodyBytes)
	log.Print(bodyString)
	io.WriteString(w, "Ok\n")
}