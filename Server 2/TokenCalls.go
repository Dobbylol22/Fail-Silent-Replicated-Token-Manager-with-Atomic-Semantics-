package main

import (
	//"context"
	//"flag"
	tk "Project_2/TokenService"
	context "context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Server struct {
	tk.UnimplementedTokenCRUDServer
	s sync.Mutex
}

type WriteData struct {
	nameValue    string
	lowValue     uint64
	midValue     uint64
	highValue    uint64
	partialValue uint64
	finalValue   uint64
}

var tokenMap = make(map[string]WriteData)
var ServerList []string

func init() {
	ServerList = []string{"8888", "8890"}
	tokenMap = make(map[string]WriteData)
	tokenIdS := ReadYaml()
	var ids []string
	for k, v := range tokenIdS {
		tokenMap[v.TokenId] = WriteData{}
		ids = append(ids, k)
	}
	log.Println("Token Ids in this sever ", ids)
}
func Hash(name string, nonce uint64) uint64 {
	hasher := sha256.New()
	hasher.Write([]byte(fmt.Sprintf("%s %d", name, nonce)))
	return binary.BigEndian.Uint64(hasher.Sum(nil))
}

func (s *Server) Create(ctx context.Context, message *tk.ID) (*tk.Status, error) {
	s.s.Lock()
	defer s.s.Unlock()
	for k := range tokenMap {
		if k == message.Body {
			log.Println("Key already exists")
			return &tk.Status{Body: "Fail"}, nil
		}
	}
	tokenMap[message.Body] = WriteData{}
	log.Println("Created Token:", message.Body)
	return &tk.Status{Body: "Success"}, nil
}

func (s *Server) Delete(ctx context.Context, message *tk.ID) (*tk.Status, error) {
	s.s.Lock()
	defer s.s.Unlock()
	Status := "Fail"
	for k := range tokenMap {
		if k == message.Body {
			delete(tokenMap, k)
			Status = "Success"
			fmt.Println("Deleted Token :", message.Body)
			break
		}
	}
	if Status == "Fail" {
		fmt.Println("Token not Found")
	}
	//log.Println("Dropping Token ID", tokenMap)
	return &tk.Status{Body: Status}, nil
}
func CheckLowMidHighForWrite(id string, low uint64, mid uint64, high uint64, name string) WriteData {
	tokens := tokenMap[id]
	if low <= mid {
		if mid < high {

			tokens.lowValue = low
			tokens.midValue = mid
			tokens.highValue = high
			tokens.nameValue = name
		}
	}
	tokenMap[id] = tokens
	return tokens
}

func CheckLowMidHighForRead(id string, low uint64, mid uint64, high uint64, name string) (WriteData, string) {
	tokens := tokenMap[id]
	if low <= mid {
		if mid < high {

			tokens.lowValue = low
			tokens.midValue = mid
			tokens.highValue = high
			tokens.nameValue = name
			tokenMap[id] = tokens
			return tokens, "success"
		}
	}
	return tokens, "failure"
}

func MinimumHashForWrite(actualLow uint64, actualMid uint64, actualName string) uint64 {
	temp := actualLow
	minHash := Hash(actualName, actualLow)
	for i := actualLow; i < actualMid; i++ {
		hashval := Hash(actualName, i)
		if hashval < minHash {
			minHash = hashval
			temp = i
		}
	}
	return temp
}

func MinimumHashForRead(actualMid uint64, actualHigh uint64, actualName string) uint64 {
	temp := actualMid
	minHash := Hash(actualName, actualMid)
	for i := actualMid; i < actualHigh; i++ {
		hashval := Hash(actualName, i)
		if hashval < minHash {
			minHash = hashval
			temp = i
		}
	}
	return temp
}

func (s *Server) Write(ctx context.Context, tokenValue *tk.Domain) (*tk.ResponseForWrite, error) {
	s.s.Lock()
	defer s.s.Unlock()
	_, haskey := tokenMap[tokenValue.Id]
	if haskey {
		finalToken := CheckLowMidHighForWrite(tokenValue.Id, tokenValue.Low, tokenValue.Mid, tokenValue.High, tokenValue.Name)
		minimumValue := MinimumHashForWrite(finalToken.lowValue, finalToken.midValue, finalToken.nameValue)
		finalToken.partialValue = minimumValue
		tokenMap[tokenValue.Id] = finalToken
		log.Println("Write Operation started.")
		fmt.Println("Token after Write Opearation:{ name :", tokenMap[tokenValue.Id].nameValue+" low: ", tokenMap[tokenValue.Id].lowValue, "mid: ", tokenMap[tokenValue.Id].midValue, "high: ", tokenMap[tokenValue.Id].highValue, " partialValue: ", tokenMap[tokenValue.Id].partialValue, "finalValue: ", tokenMap[tokenValue.Id].finalValue)
		fmt.Println()
		for i := range ServerList {

			address := fmt.Sprintf("%s:%s", "localhost", ServerList[i])
			log.Println("Replication Started in the node: ", ServerList[i])
			conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				log.Fatalf("did not connect: %s", err)
			}
			defer conn.Close()

			replica := tk.NewTokenCRUDClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			replicaObjectForWrite := tk.ReplicationDomain{
				Id:           tokenValue.Id,
				Name:         tokenMap[tokenValue.Id].nameValue,
				Low:          tokenMap[tokenValue.Id].lowValue,
				Mid:          tokenMap[tokenValue.Id].midValue,
				High:         tokenMap[tokenValue.Id].highValue,
				PartialValue: tokenMap[tokenValue.Id].partialValue,
				FinalValue:   finalToken.finalValue,
			}
			replica.ReplicationForWrite(ctx, &replicaObjectForWrite)

			defer cancel()
		}
		fmt.Println("Token after Write Opearation: ", tokenMap[tokenValue.Id])
		return &tk.ResponseForWrite{ReturnPartialValue: minimumValue}, nil
	} else {
		fmt.Println("Cannot write the token")
		return &tk.ResponseForWrite{Errored: "Cannot write the token"}, nil
	}

}
func (s *Server) ReplicationForWrite(
	ctx context.Context, requestObjectForWrite *tk.ReplicationDomain) (*tk.ReplicationResponseForWrite, error) {
	var responseForWrite tk.ReplicationResponseForWrite
	tokenMap[requestObjectForWrite.Id] = WriteData{
		nameValue:    requestObjectForWrite.Name,
		lowValue:     requestObjectForWrite.Low,
		highValue:    requestObjectForWrite.High,
		midValue:     requestObjectForWrite.Mid,
		partialValue: requestObjectForWrite.PartialValue,
		finalValue:   requestObjectForWrite.FinalValue,
	}
	log.Println("Replicated token for "+requestObjectForWrite.Id+" with values:{ name:", requestObjectForWrite.Name, " low:", requestObjectForWrite.Low, "mid:", requestObjectForWrite.Mid, "high:", requestObjectForWrite.High, " partialValue:", requestObjectForWrite.PartialValue, "finalValue:", requestObjectForWrite.FinalValue)

	responseForWrite.PartialValue = requestObjectForWrite.PartialValue
	return &responseForWrite, nil
}

func (s *Server) Read(ctx context.Context, tokenRead *tk.ID) (*tk.ResponseForRead, error) {
	s.s.Lock()
	defer s.s.Unlock()
	tokens, haskey := tokenMap[(tokenRead.Body)]
	if haskey {
		actualFinalValue := tokens.finalValue
		finalToken, msg := CheckLowMidHighForRead(tokenRead.Body, tokens.lowValue, tokens.midValue, tokens.highValue, tokens.nameValue)
		if msg == "success" {
			finalValue := MinimumHashForRead(finalToken.midValue, finalToken.highValue, finalToken.nameValue)
			if finalToken.partialValue < finalValue {
				actualFinalValue = finalToken.partialValue
			} else {
				actualFinalValue = finalValue
			}
			log.Println("Read Opearion Started.")
			////////replication for read operation//////
			for i := range ServerList {
				address := fmt.Sprintf("%s:%s", "localhost", ServerList[i])
				log.Println("Replication Started in the node: ", ServerList[i])
				conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
				if err != nil {
					log.Fatalf("did not connect: %s", err)
				}
				defer conn.Close()

				replica := tk.NewTokenCRUDClient(conn)
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				replicaObjectForRead := tk.ReplicationID{
					Body:       tokenRead.Body,
					FinalValue: actualFinalValue,
				}
				replica.ReplicationForRead(ctx, &replicaObjectForRead)
				defer cancel()
			}
			///////replication for read operation/////
		} else {
			fmt.Println("Tokens low, mid, high values are not in correct format")
			return &tk.ResponseForRead{Errored: "Cannot read the token"}, nil
		}
		fmt.Println("The Final value of token for ", tokenRead.Body, "is", actualFinalValue)
		return &tk.ResponseForRead{ReturnFinalValue: actualFinalValue}, nil
	} else {
		fmt.Println("Cannot read the token")
		return &tk.ResponseForRead{Errored: "Cannot read the token"}, nil
	}

}

func (s *Server) ReplicationForRead(
	ctx context.Context, requestObjectForRead *tk.ReplicationID) (*tk.ReplicationResponseForRead, error) {
	var responseForRead tk.ReplicationResponseForRead
	token := tokenMap[requestObjectForRead.Body]
	token.finalValue = requestObjectForRead.FinalValue
	tokenMap[requestObjectForRead.Body] = token

	log.Println("Replicated Read token "+requestObjectForRead.Body+" with Final Value: ", requestObjectForRead.FinalValue)
	//	printTokenInfo(req.Id)
	responseForRead.FinalValue = requestObjectForRead.FinalValue
	return &responseForRead, nil
}
