package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	pb "github.com/heyuhang0/DSCProject/pkg/dto"
	"github.com/montanaflynn/stats"
	"google.golang.org/grpc"
	"log"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

func main() {

	// Read command from terminal
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("$ ")
		cmdString, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		err = runCommand(cmdString)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}

func runCommand(commandStr string) error {
	commandStr = strings.TrimSuffix(commandStr, "\n")
	arrCommandStr := strings.Fields(commandStr) // split command into an array of string - Fields will separate by whitespaces

	switch arrCommandStr[0] {
	case "exit":
		os.Exit(0)

	case "get":
		if len(arrCommandStr) != 3 {
			return errors.New("GET requires 2 arguments: <address> <key>")
		}

		address, key := arrCommandStr[1], arrCommandStr[2]
		conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("Did not connect: %v", err)
		}
		defer func() { _ = conn.Close() }()
		c := pb.NewKeyValueStoreClient(conn)

		// Basic Test
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		fmt.Println("=== GET Request is called! ===")
		resp, err := c.Get(ctx, &pb.GetRequest{Key: []byte(key)})
		if err != nil {
			log.Fatalf("Failed to GET: %v", err)
		}
		fmt.Fprintln(os.Stdout, "GET SUCCESSFUL:", key, string(resp.Object))
		return nil

	case "put":
		address, key, value := arrCommandStr[1], arrCommandStr[2], arrCommandStr[3]

		conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("Did not connect: %v", err)
		}
		defer func() { _ = conn.Close() }()
		c := pb.NewKeyValueStoreClient(conn)

		// Basic Test
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		fmt.Println("=== PUT Request is called! ===")
		_, err = c.Put(ctx, &pb.PutRequest{
			Key:    []byte(key),
			Object: []byte(value),
		})
		if err != nil {
			log.Fatalf("Failed to PUT: %v", err)
		}

		log.Printf("PUT SUCCESSFUL: {%v: %v}", key, value)
		return nil

	case "rps":
		if len(arrCommandStr) != 4 {
			return errors.New("rps (request per second) requires 3 argument: <address> <key> <no_requests>")
		}
		address, key := arrCommandStr[1], arrCommandStr[2]
		noRequests, _ := strconv.Atoi(arrCommandStr[3])
		var elapsed float64

		conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("Did not connect: %v", err)
		}
		defer func() { _ = conn.Close() }()
		c := pb.NewKeyValueStoreClient(conn)

		// Basic Test
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		start := time.Now()
		for i := 0; i < noRequests; i++ {
			// start request: get or put
			_, err := c.Get(ctx, &pb.GetRequest{Key: []byte(key)})
			if err != nil {
				log.Fatalf("Failed to GET: %v", err)
			}
		}
		t := time.Now()
		elapsed = t.Sub(start).Seconds()
		requestPerSec := float64(noRequests)/(elapsed)
		fmt.Fprintln(os.Stdout, "Number of Requests Per Second:", requestPerSec)
		return nil

	case "latencytime":
		if len(arrCommandStr) != 5 {
			return errors.New("latencytime requires 4 arguments: <address> <key> <no_requests> <percentile>")
		}
		address, key := arrCommandStr[1], arrCommandStr[2]
		noRequestsStr, percentileStr := arrCommandStr[3], arrCommandStr[4]
		noRequests, _ := strconv.Atoi(noRequestsStr)
		percentileToEval, _ := strconv.ParseFloat(percentileStr, 64)

		conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("Did not connect: %v", err)
		}
		defer func() { _ = conn.Close() }()
		c := pb.NewKeyValueStoreClient(conn)

		// Basic Test
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		var latency []float64
		for i := 0; i < noRequests; i++ {
			start := time.Now()

			// start request: get or put
			_, err := c.Get(ctx, &pb.GetRequest{Key: []byte(key)})
			if err != nil {
				log.Fatalf("Failed to GET: %v", err)
			}

			t := time.Now()
			elapsed := t.Sub(start).Seconds()
			latency = append(latency, elapsed)
		}
		sort.Float64s(latency)
		result, err := stats.Percentile(latency, percentileToEval)
		fmt.Fprintln(os.Stdout, percentileToEval, "th percentile latency:", result)
		return nil
	}

	// Execute command
	cmd := exec.Command(arrCommandStr[0], arrCommandStr[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

//func grpcSetup(arrCommandStr []string) (resp, err){
//	address, key := arrCommandStr[1], arrCommandStr[2]
//	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
//	if err != nil {
//		log.Fatalf("Did not connect: %v", err)
//	}
//	defer func() { _ = conn.Close() }()
//	c := pb.NewKeyValueStoreClient(conn)
//
//	// Basic Test
//	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
//	defer cancel()
//
//	fmt.Println("=== GET Request is called! ===")
//	resp, err := c.Get(ctx, &pb.GetRequest{Key: []byte(key)})
//	return resp, err
//}