package delivery

import (
	"context"
	"errors"
	"goph_keeper/internal/server/interfaces"
	"goph_keeper/internal/server/utils"
	"goph_keeper/internal/shared/models"
	"goph_keeper/internal/shared/pb"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCTransportServer struct {
	pb.UnimplementedTransportServiceServer
	db        interfaces.RepositoryDb
	secretKey string
}

func NewGRPCHandler(db interfaces.RepositoryDb, secretKey string) *GRPCTransportServer {
	return &GRPCTransportServer{
		db:        db,
		secretKey: secretKey,
	}
}

func (s *GRPCTransportServer) Register(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	username := req.GetUsername()
	passwordHash := req.GetPasswordHash()

	if username == "" || passwordHash == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password are required")
	}

	if err := s.db.AddUser(ctx, username, passwordHash); err != nil {
		slog.Error("failed to register user in database", "username", username, "error", err)

		if errors.Is(err, interfaces.ErrUserAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "username is already taken")
		}

		return nil, status.Error(codes.Internal, "internal server error, please try again later")
	}

	token, err := utils.GenerateToken(username, s.secretKey)
	if err != nil {
		slog.Error("failed to generate token for registered user", "username", username, "error", err)
		return nil, status.Error(codes.Internal, "internal server error, please try again later")
	}

	resp := &pb.AuthResponse{}
	resp.SetToken(token)
	return resp, nil
}

func (s *GRPCTransportServer) Login(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	username := req.GetUsername()
	passwordHash := req.GetPasswordHash()

	savedHash, err := s.db.GetUserPassword(ctx, username)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	if savedHash != passwordHash {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	token, err := utils.GenerateToken(username, s.secretKey)
	if err != nil {
		slog.Error("failed to generate token for logged in user", "username", username, "error", err)
		return nil, status.Error(codes.Internal, "internal server error, please try again later")
	}

	resp := &pb.AuthResponse{}
	resp.SetToken(token)
	return resp, nil
}

func (s *GRPCTransportServer) SaveRecord(ctx context.Context, req *pb.Record) (*pb.SaveResponse, error) {
	username, ok := models.GetUserName(ctx)
	if !ok || username == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	record := models.EncryptedRecord{
		Name:     req.GetName(),
		DataType: req.GetDataType(),
		Payload:  req.GetSecureData(),
		Nonce:    req.GetNonce(),
	}

	if record.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "record name cannot be empty")
	}

	if err := s.db.SaveRecord(ctx, username, record); err != nil {
		slog.Error("saveRecord", "error", err)

		return nil, status.Error(codes.Internal, "internal server error, please try again later")
	}

	return &pb.SaveResponse{}, nil
}

func (s *GRPCTransportServer) GetRecord(ctx context.Context, req *pb.GetRecordRequest) (*pb.Record, error) {
	username, ok := models.GetUserName(ctx)
	if !ok || username == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	name := req.GetName()
	dbRecord, err := s.db.GetRecord(ctx, username, name)
	if err != nil {
		slog.Error("getRecord", "error", err)

		return nil, status.Error(codes.NotFound, "record not found")
	}

	resp := &pb.Record{}
	resp.SetName(dbRecord.Name)
	resp.SetDataType(dbRecord.DataType)
	resp.SetSecureData(dbRecord.Payload)
	resp.SetNonce(dbRecord.Nonce)

	return resp, nil
}

func (s *GRPCTransportServer) DeleteRecord(ctx context.Context, req *pb.GetRecordRequest) (*pb.SaveResponse, error) {
	username, ok := models.GetUserName(ctx)
	if !ok || username == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	if err := s.db.DeleteRecord(ctx, username, req.GetName()); err != nil {
		slog.Error("delete record", "error", err)

		return nil, status.Error(codes.Internal, "failed to delete")
	}

	return &pb.SaveResponse{}, nil
}

func (s *GRPCTransportServer) ListRecords(ctx context.Context, req *pb.ListRecordsRequest) (*pb.ListRecordsResponse, error) {
	username, ok := models.GetUserName(ctx)
	if !ok || username == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	dbMetas, err := s.db.ListRecords(ctx, username, req.GetLimit())
	if err != nil {
		slog.Error("list record", "error", err)

		return nil, status.Error(codes.Internal, "failed to fetch list")
	}

	resp := &pb.ListRecordsResponse{}

	for _, m := range dbMetas {
		meta := &pb.ListRecordsResponse_RecordMeta{}
		meta.SetName(m.Name)
		meta.SetDataType(m.DataType)

		resp.SetRecords(append(resp.GetRecords(), meta))
	}

	return resp, nil
}
