package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/AJackTi/grpc-go-calculator/calculatorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	// doServerStreaming(&c)

	// doClientStreaming(&c)

	// doBiDiStreaming(&c)

	doErrorUnary(&c)
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

func doClientStreaming(c *calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do a Client Streaming RPC...")

	stream, err := (*c).ComputeAverage(context.Background())
	if err != nil {
		log.Fatalf("error while calling ComputeAverage: %v\n", err)
	}

	// we iterate over our slice and send each message individually
	for i := 0; i < 10; i++ {
		fmt.Printf("Sending req %v\n", i)
		stream.Send(&calculatorpb.ComputeAverageRequest{
			Number: int32(i),
		})
		time.Sleep(1000 * time.Millisecond)
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("error while receiving response from ComputeAverage: %v\n", err)
	}

	fmt.Printf("ComputeAverage Response: %v\n", res.Average)
}

func doBiDiStreaming(c *calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do a BiDi Streaming RPC...")

	// we create a stream by invoking the client
	stream, err := (*c).FindMaximum(context.Background())
	if err != nil {
		log.Fatalf("Error while creating stream: %v\n", err)
	}

	waitc := make(chan struct{})
	// we send a bunch of messages to the client (go routine)
	go func() {
		// function to send a bunch of messages
		for i := 1; i < 100; i++ {
			fmt.Printf("Sending message %v\n", i)
			stream.Send(&calculatorpb.FindMaximumRequest{Number: int32(i)})
			time.Sleep(1000 * time.Millisecond)
		}

		stream.CloseSend()
	}()

	// we receive a bunch of messages from the client (go routine)
	go func() {
		// function to receive a bunch of messages
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error while receiving: %v\n", err)
				break
			}

			fmt.Printf("Received: %v\n", res.Maximum)
		}
		close(waitc)
	}()

	// block until everything is done
	<-waitc
}

func doErrorUnary(c *calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do a SquareRoot Unary RPC...")

	// correct call
	doErrorCall(c, 10)

	// error call
	doErrorCall(c, -2)
}

func doErrorCall(c *calculatorpb.CalculatorServiceClient, number int32) {
	res, err := (*c).SquareRoot(context.Background(), &calculatorpb.SquareRootRequest{Number: int32(number)})
	if err != nil {
		respErr, ok := status.FromError(err)
		if ok {
			// actual error from gRPC (user error)
			fmt.Printf("Error message from server: %v\n", respErr.Message())
			fmt.Printf("Error code from server: %v\n", respErr.Code())
			if respErr.Code() == codes.InvalidArgument {
				fmt.Println("We probably sent a negative number!")
				return
			}
		} else {
			log.Fatalf("Big Error calling SquareRoot: %v\n", err)
			return
		}
	}
	fmt.Printf("Result of square root of %v is: %v\n", number, res.GetNumberRoot())
}
