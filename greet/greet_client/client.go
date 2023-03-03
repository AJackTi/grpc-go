package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/AJackTi/grpc-go-greet/greetpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

func main() {
	fmt.Println("Hello I'm a client")

	tls := true
	opts := grpc.WithInsecure()
	if tls {
		certFile := "../../ssl/ca.crt" // Certificate Authority Trust certificate
		creds, sslErr := credentials.NewClientTLSFromFile(certFile, "")
		if sslErr != nil {
			log.Fatalf("Error while calling CA trust certificate: %v\n", sslErr)
			return
		}

		opts = grpc.WithTransportCredentials(creds)
	}

	conn, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Fatalf("could not connect: %v\n", err)
	}
	defer conn.Close()

	c := greetpb.NewGreetServiceClient(conn)
	// fmt.Printf("Created client: %f", c)

	doUnary(&c)

	// doServerStreaming(&c)

	// doClientStreaming(&c)

	// doBiDiStreaming(&c)

	// doUnaryWithDeadline(&c, 5*time.Second) // should complete
	// doUnaryWithDeadline(&c, 1*time.Second) // should timeout
}

func doUnary(c *greetpb.GreetServiceClient) {
	fmt.Println("Starting to do a Unary RPC...")
	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Ti",
			LastName:  "AJack",
		},
	}
	res, err := (*c).Greet(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling Greet RPC: %v\n", err)
	}

	log.Printf("Response from Greet: %v\n", res.Result)
}

func doServerStreaming(c *greetpb.GreetServiceClient) {
	fmt.Println("Starting to do a Server Streaming RPC...")

	req := &greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Ti",
			LastName:  "AJack",
		},
	}

	resStream, err := (*c).GreetManyTimes(context.Background(), req)
	if err != nil {
		log.Fatalf("error while calling GreetManyTimes RPC: %v\n", err)
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
		log.Printf("Response from GreetManyTimes: %v\n", msg.GetResult())
	}
}

func initDummyData() []*greetpb.LongGreetRequest {
	response := make([]*greetpb.LongGreetRequest, 0)
	for i := 0; i < 10; i++ {
		response = append(response, &greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: fmt.Sprint(i * i),
				LastName:  fmt.Sprint(i * i),
			},
		})
	}

	return response
}

func doClientStreaming(c *greetpb.GreetServiceClient) {
	fmt.Println("Starting to do a Client Streaming RPC...")

	stream, err := (*c).LongGreet(context.Background())
	if err != nil {
		log.Fatalf("error while calling LongGreet: %v\n", err)
	}

	// we iterate over our slice and send each message individually
	for _, data := range initDummyData() {
		fmt.Printf("Sending req %v\n", data)
		stream.Send(data)
		time.Sleep(1000 * time.Millisecond)
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("error while receiving response from LongGreet: %v\n", err)
	}

	fmt.Printf("LongGreet Response: %v\n", res.Result)
}

func initDummyGreetEveryOneData() []*greetpb.GreetEveryoneRequest {
	response := make([]*greetpb.GreetEveryoneRequest, 0)
	for i := 0; i < 10; i++ {
		response = append(response, &greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: fmt.Sprint(i * i),
				LastName:  fmt.Sprint(i * i),
			},
		})
	}

	return response
}

func doBiDiStreaming(c *greetpb.GreetServiceClient) {
	fmt.Println("Starting to do a BiDi Streaming RPC...")

	// we create a stream by invoking the client
	stream, err := (*c).GreetEveryone(context.Background())
	if err != nil {
		log.Fatalf("Error while creating stream: %v\n", err)
	}

	waitc := make(chan struct{})
	// we send a bunch of messages to the client (go routine)
	go func() {
		// function to send a bunch of messages
		for _, data := range initDummyGreetEveryOneData() {
			fmt.Printf("Sending message %v\n", data)
			stream.Send(data)
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

			fmt.Printf("Received: %v\n", res.GetResult())
		}
		close(waitc)
	}()

	// block until everything is done
	<-waitc
}

func doUnaryWithDeadline(c *greetpb.GreetServiceClient, timeout time.Duration) {
	fmt.Println("Starting to do a UnaryWithDeadline RPC...")
	req := &greetpb.GreetWithDeadlineRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Ti",
			LastName:  "AJack",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, err := (*c).GreetWithDeadline(ctx, req)
	if err != nil {
		statusErr, ok := status.FromError(err)
		if ok {
			if statusErr.Code() == codes.DeadlineExceeded {
				fmt.Println("Timeout was hit! Deadline was exceeded")
			} else {
				fmt.Printf("unexpected error: %v\n", statusErr.Message())
			}
		} else {
			log.Fatalf("error while calling GreetWithDeadline: %v\n", err)
		}
		return
	}

	log.Printf("Response from GreetWithDeadline: %v\n", res)
}
