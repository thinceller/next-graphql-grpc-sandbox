package main

import (
	"context"
	"log"
	"net"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/protobuf/ptypes"
	"github.com/jmoiron/sqlx"
	pb "github.com/thinceller/next-graphql-grpc-sandbox/todo-app/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	port = ":9101"
)

type Todo struct {
	Id        int32     `db:"id"`
	Title     string    `db:"title"`
	Content   string    `db:"content"`
	Done      bool      `db:"done"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type server struct {
	pb.UnimplementedTodoServiceServer
	db *sqlx.DB
}

func (s server) GetTodoList(ctx context.Context, _ *emptypb.Empty) (*pb.TodoListResponse, error) {
	log.Print("todo.TodoService/GetTodoList")

	var dbtodos []*Todo
	var todos []*pb.Todo
	query := "SELECT id,title,done,created_at,updated_at FROM todos ORDER BY updated_at DESC"
	if err := s.db.Select(&dbtodos, query); err != nil {
		return nil, err
	}

	for _, t := range dbtodos {
		created_at, _ := ptypes.TimestampProto(t.CreatedAt)
		updated_at, _ := ptypes.TimestampProto(t.UpdatedAt)

		todos = append(todos, &pb.Todo{
			Id:        t.Id,
			Title:     t.Title,
			Done:      t.Done,
			CreatedAt: created_at,
			UpdatedAt: updated_at,
		})
	}

	return &pb.TodoListResponse{
		Todos: todos,
	}, nil
}

func (s server) GetTodoDetail(ctx context.Context, req *pb.TodoDetailRequest) (*pb.TodoDetailResponse, error) {
	log.Print("todo.TodoService/GetTodoDetail")

	var dbtodo Todo
	query := "SELECT id,title,content,done,created_at,updated_at FROM todos WHERE id = ?"
	id := strconv.Itoa(int(req.GetId()))
	if err := s.db.Get(&dbtodo, query, id); err != nil {
		return nil, err
	}

	created_at, _ := ptypes.TimestampProto(dbtodo.CreatedAt)
	updated_at, _ := ptypes.TimestampProto(dbtodo.UpdatedAt)
	todo := pb.TodoDetailResponse{
		Id:        dbtodo.Id,
		Title:     dbtodo.Title,
		Content:   dbtodo.Content,
		Done:      dbtodo.Done,
		CreatedAt: created_at,
		UpdatedAt: updated_at,
	}

	return &todo, nil
}

func (s server) CreateTodo(ctx context.Context, req *pb.CreateTodoRequest) (*pb.TodoDetailResponse, error) {
	log.Print("todo.TodoService/CreateTodo")

	query := "INSERT INTO todos (title,content,done) VALUES (?,?,?)"
	title := req.GetTitle()
	content := req.GetContent()
	done := req.GetDone()
	result, err := s.db.Exec(query, title, content, done)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	var dbtodo Todo
	query = "SELECT id,title,content,done,created_at,updated_at FROM todos WHERE id = ?"
	if err := s.db.Get(&dbtodo, query, id); err != nil {
		return nil, err
	}

	created_at, _ := ptypes.TimestampProto(dbtodo.CreatedAt)
	updated_at, _ := ptypes.TimestampProto(dbtodo.UpdatedAt)
	todo := pb.TodoDetailResponse{
		Id:        dbtodo.Id,
		Title:     dbtodo.Title,
		Content:   dbtodo.Content,
		Done:      dbtodo.Done,
		CreatedAt: created_at,
		UpdatedAt: updated_at,
	}

	return &todo, nil
}

func (s server) UpdateTodo(ctx context.Context, req *pb.TodoDetail) (*pb.TodoDetailResponse, error) {
	log.Print("todo.TodoService/UpdateTodo")

	query := "UPDATE todos SET title = ?, content = ?, done = ? where id = ?"
	id := strconv.Itoa(int(req.GetId()))
	title := req.GetTitle()
	content := req.GetContent()
	done := req.GetDone()
	if _, err := s.db.Exec(query, title, content, done, id); err != nil {
		return nil, err
	}

	var dbtodo Todo
	query = "SELECT id,title,content,done,created_at,updated_at FROM todos WHERE id = ?"
	if err := s.db.Get(&dbtodo, query, id); err != nil {
		return nil, err
	}

	created_at, _ := ptypes.TimestampProto(dbtodo.CreatedAt)
	updated_at, _ := ptypes.TimestampProto(dbtodo.UpdatedAt)
	todo := pb.TodoDetailResponse{
		Id:        dbtodo.Id,
		Title:     dbtodo.Title,
		Content:   dbtodo.Content,
		Done:      dbtodo.Done,
		CreatedAt: created_at,
		UpdatedAt: updated_at,
	}

	return &todo, nil
}

func (s server) DeleteTodo(ctx context.Context, req *pb.TodoDetailRequest) (*emptypb.Empty, error) {
	log.Print("todo.TodoService/GetTodoDetail")

	query := "DELETE FROM todos where id = ?"
	id := strconv.Itoa(int(req.GetId()))
	if _, err := s.db.Exec(query, id); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}

	// establish db connection
	db, err := sqlx.Open("mysql", "root:password@tcp(127.0.0.1:9000)/todo?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	pb.RegisterTodoServiceServer(s, &server{db: db})
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
