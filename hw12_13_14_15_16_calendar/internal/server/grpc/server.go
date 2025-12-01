package grpcserver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/api/event"
	"github.com/AnastasiaDAmber/golang_homework/hw12_13_14_15_calendar/internal/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	event.UnimplementedEventServiceServer
	logger  Logger
	app     Application
	host    string
	port    int
	grpcSrv *grpc.Server
}

type Logger interface {
	Info(msg string)
	Error(msg string)
	Debug(msg string)
}

type Application interface {
	CreateEvent(ctx context.Context, e storage.Event) error
	UpdateEvent(ctx context.Context, e storage.Event) error
	DeleteEvent(ctx context.Context, id string) error
	GetEvent(ctx context.Context, id string) (storage.Event, error)
	ListEvents(ctx context.Context) ([]storage.Event, error)
	ListEventsDay(ctx context.Context, dayStart time.Time) ([]storage.Event, error)
	ListEventsWeek(ctx context.Context, weekStart time.Time) ([]storage.Event, error)
	ListEventsMonth(ctx context.Context, monthStart time.Time) ([]storage.Event, error)
}

func NewServer(logger Logger, app Application, host string, port int) *Server {
	s := &Server{
		logger: logger,
		app:    app,
		host:   host,
		port:   port,
	}

	grpcSrv := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor(logger)),
	)
	event.RegisterEventServiceServer(grpcSrv, s)
	// Включаем reflection для grpcurl
	reflection.Register(grpcSrv)
	s.grpcSrv = grpcSrv

	return s
}

func (s *Server) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.logger.Info("grpc server listening on " + addr)

	go func() {
		if err := s.grpcSrv.Serve(ln); err != nil {
			s.logger.Error("grpc serve error: " + err.Error())
		}
	}()

	<-ctx.Done()
	return s.Stop(context.Background())
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("shutting down grpc server")
	stopped := make(chan struct{})
	go func() {
		s.grpcSrv.GracefulStop()
		close(stopped)
	}()

	select {
	case <-ctx.Done():
		s.grpcSrv.Stop()
		return ctx.Err()
	case <-stopped:
		return nil
	}
}

// loggingInterceptor логирует каждый GRPC запрос
func loggingInterceptor(logger Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		latency := time.Since(start)

		statusCode := codes.OK
		if err != nil {
			if st, ok := status.FromError(err); ok {
				statusCode = st.Code()
			} else {
				statusCode = codes.Internal
			}
		}

		logger.Info(fmt.Sprintf("GRPC %s - %s - %v - %v", info.FullMethod, statusCode, latency, err))
		return resp, err
	}
}

// Конвертация между proto и доменными типами

func protoEventToDomain(pb *event.Event) (storage.Event, error) {
	e := storage.Event{
		ID:          pb.GetId(),
		Title:       pb.GetTitle(),
		Description: pb.GetDescription(),
		UserID:      pb.GetUserId(),
	}

	if pb.GetAt() != nil {
		e.At = pb.GetAt().AsTime()
	}

	if pb.GetDuration() != nil {
		e.Duration = pb.GetDuration().AsDuration()
	}

	if pb.GetNotifyBefore() != nil {
		e.NotifyBefore = pb.GetNotifyBefore().AsDuration()
	}

	return e, nil
}

func domainEventToProto(e storage.Event) *event.Event {
	pb := &event.Event{
		Id:          e.ID,
		Title:       e.Title,
		Description: e.Description,
		UserId:      e.UserID,
	}

	if !e.At.IsZero() {
		pb.At = timestamppb.New(e.At)
	}

	if e.Duration != 0 {
		pb.Duration = durationpb.New(e.Duration)
	}

	if e.NotifyBefore != 0 {
		pb.NotifyBefore = durationpb.New(e.NotifyBefore)
	}

	return pb
}

// GRPC методы

func (s *Server) CreateEvent(ctx context.Context, req *event.CreateEventRequest) (*event.CreateEventResponse, error) {
	if req.GetEvent() == nil {
		return nil, status.Error(codes.InvalidArgument, "event is required")
	}

	domainEvent, err := protoEventToDomain(req.GetEvent())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := s.app.CreateEvent(ctx, domainEvent); err != nil {
		if errors.Is(err, storage.ErrDateBusy) {
			return nil, status.Error(codes.AlreadyExists, "event with this ID already exists")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &event.CreateEventResponse{Id: domainEvent.ID}, nil
}

func (s *Server) UpdateEvent(ctx context.Context, req *event.UpdateEventRequest) (*event.UpdateEventResponse, error) {
	if req.GetEvent() == nil {
		return nil, status.Error(codes.InvalidArgument, "event is required")
	}

	domainEvent, err := protoEventToDomain(req.GetEvent())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := s.app.UpdateEvent(ctx, domainEvent); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "event not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &event.UpdateEventResponse{Success: true}, nil
}

func (s *Server) DeleteEvent(ctx context.Context, req *event.DeleteEventRequest) (*event.DeleteEventResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if err := s.app.DeleteEvent(ctx, req.GetId()); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "event not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &event.DeleteEventResponse{Success: true}, nil
}

func (s *Server) GetEvent(ctx context.Context, req *event.GetEventRequest) (*event.GetEventResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	domainEvent, err := s.app.GetEvent(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "event not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &event.GetEventResponse{Event: domainEventToProto(domainEvent)}, nil
}

func (s *Server) ListEvents(ctx context.Context, req *event.ListEventsRequest) (*event.ListEventsResponse, error) {
	events, err := s.app.ListEvents(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbEvents := make([]*event.Event, 0, len(events))
	for _, e := range events {
		pbEvents = append(pbEvents, domainEventToProto(e))
	}

	return &event.ListEventsResponse{Events: pbEvents}, nil
}

func (s *Server) ListEventsDay(ctx context.Context, req *event.ListEventsDayRequest) (*event.ListEventsDayResponse, error) {
	if req.GetDayStart() == nil {
		return nil, status.Error(codes.InvalidArgument, "day_start is required")
	}

	dayStart := req.GetDayStart().AsTime()
	events, err := s.app.ListEventsDay(ctx, dayStart)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbEvents := make([]*event.Event, 0, len(events))
	for _, e := range events {
		pbEvents = append(pbEvents, domainEventToProto(e))
	}

	return &event.ListEventsDayResponse{Events: pbEvents}, nil
}

func (s *Server) ListEventsWeek(ctx context.Context, req *event.ListEventsWeekRequest) (*event.ListEventsWeekResponse, error) {
	if req.GetWeekStart() == nil {
		return nil, status.Error(codes.InvalidArgument, "week_start is required")
	}

	weekStart := req.GetWeekStart().AsTime()
	events, err := s.app.ListEventsWeek(ctx, weekStart)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbEvents := make([]*event.Event, 0, len(events))
	for _, e := range events {
		pbEvents = append(pbEvents, domainEventToProto(e))
	}

	return &event.ListEventsWeekResponse{Events: pbEvents}, nil
}

func (s *Server) ListEventsMonth(ctx context.Context, req *event.ListEventsMonthRequest) (*event.ListEventsMonthResponse, error) {
	if req.GetMonthStart() == nil {
		return nil, status.Error(codes.InvalidArgument, "month_start is required")
	}

	monthStart := req.GetMonthStart().AsTime()
	events, err := s.app.ListEventsMonth(ctx, monthStart)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbEvents := make([]*event.Event, 0, len(events))
	for _, e := range events {
		pbEvents = append(pbEvents, domainEventToProto(e))
	}

	return &event.ListEventsMonthResponse{Events: pbEvents}, nil
}
