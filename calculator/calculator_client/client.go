package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/AJackTi/grpc-go-calculator/calculatorpb"
	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Hello I'm a client")
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect: %v\n", err)
	}
	defer conn.Close()

	c := calculatorpb.NewCalculatorServiceClient(conn)

	// doUnary(&c)

	doServerStreaming(&c)
}

func doUnary(c *calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do a Unary RPC...")
	req := &calculatorpb.SumRequest{
		FirstNumber:  10,
		SecondNumber: 12,
	}
	res, err := (*c).Sum(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling Calculator RPC: %v\n", err)
	}

	log.Printf("Response from Calculator: %v\n", res.SumResult)
}

func doServerStreaming(c *calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do a Server Streaming RPC...")

	req := &calculatorpb.PrimeNumberDecompositionRequest{
		Number: 100,
	}

	resStream, err := (*c).PrimeNumberDecomposition(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling PrimeNumberDecomposition RPC: %v\n", err)
	}

	for {
		msg, err := resStream.Recv()
		if err == io.EOF {
			// we've reached the end of the stream
			break
		}
		if err != nil {
			log.Fatalf("error while reading stream: %v\n", err)
		}
		log.Printf("Response from PrimeNumberDecomposition: %v\n", msg.PrimeFactor)
	}
}
