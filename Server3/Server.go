package main

import (
	tk "Project_2/TokenService"
	"fmt"

	"flag"

	"log"
	"net"

	"google.golang.org/grpc"
)

func main() {
	portNumber := flag.String("port", "9000", "a string")
	hostName := flag.String("host", "localhost", "a String")

	flag.Parse()
	lis, err := net.Listen("tcp", *hostName+":"+*portNumber)

	if err != nil {
		log.Fatalf("Failed to listen on %s  %s", *portNumber, err)
	} else {
		fmt.Println("Server Started at port: ", *portNumber)
	}

	s := Server{}
	grpcServer := grpc.NewServer()
	tk.RegisterTokenCRUDServer(grpcServer, &s)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Cannot serve on this port %s %s", *portNumber, err)

	}

}
