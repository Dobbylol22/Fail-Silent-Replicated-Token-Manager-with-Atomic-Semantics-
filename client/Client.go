package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	tk "Project_2/TokenService"

	"google.golang.org/grpc"
)

func main() {
	var conn *grpc.ClientConn
	functionality := flag.String("function", "create", "a string")
	tokenID := flag.Int("ID", 1234, "an int")
	//hostName := flag.String("host", "localhost", "a String")
	//portNumber := flag.String("port", "9000", "a string")
	name := flag.String("name", "abcd", "a string")
	lowNumber := flag.Uint64("low", 0000, "an int")
	midNumber := flag.Uint64("mid", 0000, "an int")
	highNumber := flag.Uint64("high", 0000, "an int")
	flag.Parse()

	tokenIdS := ReadYaml()
	//log.Println(tokenIdS)
	stringer := strconv.Itoa(*tokenID)
	WriterNode := tokenIdS[stringer].Writer
	//fmt.Println(WriterNode)
	ReaderPorts := tokenIdS[stringer].Reader
	x1 := rand.NewSource(time.Now().UnixNano())
	y1 := rand.New(x1)
	randomInt := y1.Intn(3)
	ReaderNode := ReaderPorts[randomInt]
	conn, err := grpc.Dial("localhost:8888", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Error")
	}
	defer conn.Close()
	c := tk.NewTokenCRUDClient(conn)
	Id := strconv.Itoa(*tokenID)
	actualFucntionalityName := strings.ToLower(*functionality)
	switch actualFucntionalityName {
	case "create":
		response1, err := c.Create(context.Background(), &tk.ID{Body: Id})
		if err == nil {
			log.Printf("Create Response from server %s", response1.Body)
		}
	case "delete":
		response2, err := c.Delete(context.Background(), &tk.ID{Body: Id})
		if err == nil {
			log.Printf("Delete Response from server %s", response2.Body)
		}
	case "write":
		conn, err := grpc.Dial("localhost"+":"+WriterNode, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("Error")
		}
		defer conn.Close()
		c := tk.NewTokenCRUDClient(conn)
		response3, err := c.Write(context.Background(), &tk.Domain{Id: Id, Name: *name, Low: *lowNumber, Mid: *midNumber, High: *highNumber})
		if err == nil {
			if len(response3.GetErrored()) == 0 {
				fmt.Println("Write function executed successfully with partial value:", response3.GetReturnPartialValue())
			} else {
				fmt.Println(response3.GetErrored())
			}
		}
	case "read":
		conn, err := grpc.Dial("localhost"+":"+ReaderNode, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("Error")
		}
		defer conn.Close()
		c := tk.NewTokenCRUDClient(conn)
		response4, err := c.Read(context.Background(), &tk.ID{Body: Id})
		if err == nil {
			if len(response4.GetErrored()) == 0 {
				fmt.Println("Read function executed successfully with final value:", response4.GetReturnFinalValue())
			} else {
				fmt.Println(response4.GetErrored())
			}
		}
	default:
		fmt.Println("Wrong Functionality Name specified")

	}
}
