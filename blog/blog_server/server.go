package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/AJackTi/grpc-go-blog/blogpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var collection *mongo.Collection

type server struct {
	blogpb.UnimplementedBlogServiceServer
}

type blogItem struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	AuthorID string             `bson:"author_id"`
	Content  string             `bson:"content"`
	Title    string             `bson:"title"`
}

func (b *blogItem) dataToBlogPb() *blogpb.Blog {
	return &blogpb.Blog{
		Id:       b.ID.Hex(),
		AuthorId: b.AuthorID,
		Title:    b.Title,
		Content:  b.Content,
	}
}

func (*server) CreateBlog(ctx context.Context, req *blogpb.CreateBlogRequest) (*blogpb.CreateBlogResponse, error) {
	fmt.Println("Create blog request")
	blog := req.GetBlog()

	data := blogItem{
		AuthorID: blog.GetAuthorId(),
		Content:  blog.GetContent(),
		Title:    blog.GetTitle(),
	}

	res, err := collection.InsertOne(context.Background(), data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal error: %v\n", err))
	}

	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(codes.Internal, "Cannot convert to OID")
	}

	blog.AuthorId = oid.Hex()
	return &blogpb.CreateBlogResponse{Blog: blog}, nil
}

func (*server) ReadBlog(ctx context.Context, req *blogpb.ReadBlogRequest) (*blogpb.ReadBlogResponse, error) {
	fmt.Println("Read blog request")
	blogID := req.GetBlogId()
	oid, err := primitive.ObjectIDFromHex(blogID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Cannot parse ID")
	}

	// create an empty struct
	data := &blogItem{}
	res := collection.FindOne(context.Background(), bson.M{"_id": oid})
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Cannot find blog with specified ID: %v\n", req.GetBlogId()))
	}

	return &blogpb.ReadBlogResponse{Blog: data.dataToBlogPb()}, nil
}

func (*server) UpdateBlog(ctx context.Context, req *blogpb.UpdateBlogRequest) (*blogpb.UpdateBlogResponse, error) {
	fmt.Println("Update blog request")
	blogID := req.Blog.GetId()
	oid, err := primitive.ObjectIDFromHex(blogID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Cannot parse ID")
	}

	// create an empty struct
	data := &blogItem{}

	filter := bson.D{{Key: "_id", Value: oid}}
	update := bson.D{{Key: "$set", Value: blogItem{
		AuthorID: req.Blog.GetAuthorId(),
		Title:    req.Blog.GetTitle(),
		Content:  req.Blog.GetContent(),
	}}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	res := collection.FindOneAndUpdate(context.Background(), filter, update, opts)
	if err := res.Decode(data); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Cannot find blog with specified ID: %v\n", req.Blog.GetId()))
		}

		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Cannot update object in MongoDB: %v\n", err))
	}

	return &blogpb.UpdateBlogResponse{Blog: data.dataToBlogPb()}, nil
}

func (*server) DeleteBlog(ctx context.Context, req *blogpb.DeleteBlogRequest) (*blogpb.DeleteBlogResponse, error) {
	fmt.Println("Delete blog request")
	blogID := req.GetBlogId()
	oid, err := primitive.ObjectIDFromHex(blogID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Cannot parse ID")
	}

	filter := bson.D{{Key: "_id", Value: oid}}
	res, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Cannot delete object in MongoDB: %v\n", err))
	}

	if res.DeletedCount == 0 {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Cannot find blog with specified ID: %v\n", req.GetBlogId()))
	}

	return &blogpb.DeleteBlogResponse{BlogId: req.GetBlogId()}, nil
}

func (*server) ListBlog(req *blogpb.ListBlogRequest, stream blogpb.BlogService_ListBlogServer) error {
	fmt.Println("List blog request")

	cur, err := collection.Find(context.Background(), bson.D{{}})
	if err != nil {
		return status.Errorf(codes.Internal, fmt.Sprintf("Unknown internal error: %v\n", err))
	}
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		data := &blogItem{}
		err := cur.Decode(data)
		if err != nil {
			return status.Errorf(codes.Internal, fmt.Sprintf("Error while decoding data from MongoDB: %v\n", err))
		}
		stream.Send(&blogpb.ListBlogResponse{
			Blog: data.dataToBlogPb(),
		})
	}

	if err := cur.Err(); err != nil {
		return status.Errorf(codes.Internal, fmt.Sprintf("Unknown internal error: %v\n", err))
	}

	return nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Println("Connecting to MongoDB")
	// connect to mongo
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}

	collection = (*mongo.Collection)(client.Database("mydb").Collection("blog"))

	fmt.Println("Blog Service started")
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{}
	s := grpc.NewServer(opts...)
	blogpb.RegisterBlogServiceServer(s, &server{})

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v\n", err)
		}
	}()

	// Wait for Control C to exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	// Block until a signal is received
	<-ch
	fmt.Println("Stopping the server")
	s.Stop()
	fmt.Println("Closing the listener")
	lis.Close()
	fmt.Println("Closing MongoDB Connection")
	client.Disconnect(context.TODO())
	fmt.Println("End of Program")
}
