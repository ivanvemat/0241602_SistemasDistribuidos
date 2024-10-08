package main

import (
	"context"
	"fmt"
	"net"
	"os"
	api "server/api/v1"
	log "server/log"

	"google.golang.org/grpc"
)

var _ api.LogServer = (*grpcServer)(nil)

type Config struct {
	CommitLog *log.Log
}

type grpcServer struct {
	api.UnimplementedLogServer
	CommitLog *log.Log
}

func NewGRPCServer(config *Config) (srv *grpc.Server, err error) {
	server := &grpcServer{
		CommitLog: config.CommitLog,
	}
	var opts []grpc.ServerOption
	srv = grpc.NewServer(opts...)
	api.RegisterLogServer(srv, server)
	return srv, nil
}

func (s *grpcServer) Produce(ctx context.Context, req *api.ProduceRequest) (*api.ProduceResponse, error) {
	offset, err := s.CommitLog.Append(req.Record)
	if err != nil {
		return nil, err
	}
	return &api.ProduceResponse{Offset: offset}, nil
}

func (s *grpcServer) Consume(ctx context.Context, req *api.ConsumeRequest) (*api.ConsumeResponse, error) {
	record, err := s.CommitLog.Read(req.Offset)
	if err != nil {
		outOfRange, ok := err.(*api.ErrOffsetOutOfRange)
		if ok {
			return nil, outOfRange.GRPCStatus().Err()
		}
		return nil, err
	}
	return &api.ConsumeResponse{Record: record}, nil
}

func (s *grpcServer) ProduceStream(stream api.Log_ProduceStreamServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}
		res, err := s.Produce(stream.Context(), req)
		if err != nil {
			return err
		}
		if err = stream.Send(res); err != nil {
			return err
		}
	}
}

func (s *grpcServer) ConsumeStream(req *api.ConsumeRequest, stream api.Log_ConsumeStreamServer) error {
	for {
		select {
		case <-stream.Context().Done():
			return nil
		default:
			res, err := s.Consume(stream.Context(), req)
			switch err.(type) {
			case nil:
			case api.ErrOffsetOutOfRange:
				continue
			default:
				return err
			}
			if err = stream.Send(res); err != nil {
				return err
			}
			req.Offset++
		}
	}
} 

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println(err)
		return
	}

	dir, err := os.MkdirTemp("..", "log_data")
	if err != nil {
		fmt.Println(err)
		return
	}

	server := grpc.NewServer()
	commitLog, err := log.NewLog(dir, log.Config{})
	if err != nil {
		fmt.Println(err)
		return
	}

	api.RegisterLogServer(server, &grpcServer{CommitLog: commitLog})
	server.Serve(listener)
}