package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/AJackTi/grpc-go-blog/blogpb"
	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Blog Client")

	opts := grpc.WithInsecure()

	conn, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Fatalf("could not connect: %v\n", err)
	}
	defer conn.Close()

	c := blogpb.NewBlogServiceClient(conn)

	// createBlog(&c)

	// readBlog(&c, "6402b79e1bc075bca52a4137")

	// updateBlog(&c, "6402b79e1bc075bca52a4137")

	// deleteBlog(&c, "6402b79e1bc075bca52a4131")

	listBlog(&c)
}

func createBlog(c *blogpb.BlogServiceClient) {
	fmt.Println("Creating the blog")
	req := &blogpb.CreateBlogRequest{
		Blog: &blogpb.Blog{
			AuthorId: "AJackTi",
			Title:    fmt.Sprintf("Learning Golang %d", time.Now().Nanosecond()),
			Content:  fmt.Sprintf("Golang %d", time.Now().Nanosecond()),
		},
	}
	res, err := (*c).CreateBlog(context.Background(), req)
	if err != nil {
		log.Fatalf("Unexpected error: %v\n", err)
	}

	log.Printf("Blog has been created: %v\n", res.GetBlog())
}

func readBlog(c *blogpb.BlogServiceClient, id string) {
	fmt.Println("Reading the blog")
	req := &blogpb.ReadBlogRequest{
		BlogId: id,
	}
	res, err := (*c).ReadBlog(context.Background(), req)
	if err != nil {
		log.Fatalf("Unexpected error: %v\n", err)
	}

	log.Printf("Response from ReadBlog: %v\n", res.GetBlog())
}

func updateBlog(c *blogpb.BlogServiceClient, id string) {
	fmt.Println("Updating the blog")
	req := &blogpb.UpdateBlogRequest{
		Blog: &blogpb.Blog{
			Id:       id,
			AuthorId: "Ti 2",
			Title:    "Golang Updated 2",
			Content:  "Golang Updated 2",
		},
	}
	res, err := (*c).UpdateBlog(context.Background(), req)
	if err != nil {
		log.Fatalf("Unexpected error: %v\n", err)
	}

	log.Printf("Response from UpdateBlog: %v\n", res.GetBlog())
}

func deleteBlog(c *blogpb.BlogServiceClient, id string) {
	fmt.Println("Deleting the blog")
	req := &blogpb.DeleteBlogRequest{
		BlogId: id,
	}
	res, err := (*c).DeleteBlog(context.Background(), req)
	if err != nil {
		log.Fatalf("Unexpected error: %v\n", err)
	}

	log.Printf("Response from DeleteBlog: %v\n", res.BlogId)
}

func listBlog(c *blogpb.BlogServiceClient) {
	fmt.Println("Listing the blog")
	req := &blogpb.ListBlogRequest{}
	resStream, err := (*c).ListBlog(context.Background(), req)
	if err != nil {
		log.Fatalf("Unexpected error: %v\n", err)
	}

	for {
		res, err := resStream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("error while reading stream: %v\n", err)
		}
		log.Printf("Response from ListBlog: %v\n", res.Blog)
	}
}
