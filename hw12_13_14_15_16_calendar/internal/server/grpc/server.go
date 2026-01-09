package grpcserver

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/configuration"
	"github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/pb"
	srvcontr "github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/server/contracts"
	storagecontracts "github.com/MaximBayurov/otus-go-hw/hw12_13_14_15_calendar/internal/storage/contracts"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedCalendarServer

	configs configuration.ServerConf
	logger  srvcontr.Logger
	app     srvcontr.Application
}

func NewServer(
	logger srvcontr.Logger,
	app srvcontr.Application,
	configs configuration.ServerConf,
) *Server {
	return &Server{
		logger:  logger,
		app:     app,
		configs: configs,
	}
}

func (s *Server) Start(_ context.Context) error {
	lsn, err := net.Listen("tcp", fmt.Sprintf(":%d", s.configs.Port)) //nolint:noctx
	if err != nil {
		log.Fatal(err)
	}

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			UnaryLogRequestInterceptor(s.logger),
		),
	)
	pb.RegisterCalendarServer(server, s)

	s.logger.Info(fmt.Sprintf("Starting gRPC server on %s", lsn.Addr().String()))

	return server.Serve(lsn)
}

func (s *Server) Stop(_ context.Context) error {
	return nil
}

func (s *Server) CreateEvent(ctx context.Context, request *pb.CreateEventRequest) (*pb.CreateEventResponse, error) {
	var event storagecontracts.Event
	var err error
	if event, err = s.app.CreateEvent(
		ctx,
		request.Event.GetTitle(),
		request.Event.GetStartTime().AsTime(),
		request.Event.GetEndTime().AsTime(),
		request.Event.GetDesc(),
		request.Event.GetOwnerId(),
		request.Event.GetNotifyTime().AsTime(),
	); err != nil {
		return nil, err
	}
	return &pb.CreateEventResponse{
		Event: &pb.Event{
			Id:         event.ID,
			Title:      event.Title,
			StartTime:  timestamppb.New(event.From),
			EndTime:    timestamppb.New(event.To),
			Desc:       event.Description,
			OwnerId:    event.OwnerID,
			NotifyTime: timestamppb.New(event.Notify),
		},
	}, nil
}

func (s *Server) UpdateEvent(ctx context.Context, request *pb.UpdateEventRequest) (*pb.UpdateEventResponse, error) {
	var event storagecontracts.Event
	var err error
	if event, err = s.app.UpdateEvent(
		ctx,
		request.EventId,
		request.UpdatedEvent.GetTitle(),
		request.UpdatedEvent.GetStartTime().AsTime(),
		request.UpdatedEvent.GetEndTime().AsTime(),
		request.UpdatedEvent.GetDesc(),
		request.UpdatedEvent.GetNotifyTime().AsTime(),
	); err != nil {
		return nil, err
	}
	return &pb.UpdateEventResponse{
		Event: &pb.Event{
			Id:         event.ID,
			Title:      event.Title,
			StartTime:  timestamppb.New(event.From),
			EndTime:    timestamppb.New(event.To),
			Desc:       event.Description,
			OwnerId:    event.OwnerID,
			NotifyTime: timestamppb.New(event.Notify),
		},
	}, nil
}

func (s *Server) DeleteEvent(ctx context.Context, request *pb.DeleteEventRequest) (*pb.DeleteEventResponse, error) {
	if err := s.app.DeleteEvent(
		ctx,
		request.EventId,
	); err != nil {
		return nil, err
	}
	return &pb.DeleteEventResponse{}, nil
}

func (s *Server) GetEventsForDay(
	ctx context.Context,
	request *pb.GetEventsForDayRequest,
) (*pb.GetEventsForDayResponse, error) {
	var events []storagecontracts.Event
	var err error
	if events, err = s.app.GetEventsForDay(ctx, request.Day.AsTime()); err != nil {
		return nil, err
	}
	return &pb.GetEventsForDayResponse{Events: s.formatEvents(events)}, nil
}

func (s *Server) GetEventsForWeek(
	ctx context.Context,
	request *pb.GetEventsForWeekRequest,
) (*pb.GetEventsForWeekResponse, error) {
	var events []storagecontracts.Event
	var err error
	if events, err = s.app.GetEventsForDay(ctx, request.StartOfWeek.AsTime()); err != nil {
		return nil, err
	}
	return &pb.GetEventsForWeekResponse{Events: s.formatEvents(events)}, nil
}

func (s *Server) GetEventsForMonth(
	ctx context.Context,
	request *pb.GetEventsForMonthRequest,
) (*pb.GetEventsForMonthResponse, error) {
	var events []storagecontracts.Event
	var err error
	if events, err = s.app.GetEventsForDay(ctx, request.StartOfMonth.AsTime()); err != nil {
		return nil, err
	}
	return &pb.GetEventsForMonthResponse{Events: s.formatEvents(events)}, nil
}

func (s *Server) formatEvents(events []storagecontracts.Event) []*pb.Event {
	result := make([]*pb.Event, len(events))
	for i, event := range events {
		result[i] = &pb.Event{
			Id:         event.ID,
			Title:      event.Title,
			StartTime:  timestamppb.New(event.From),
			EndTime:    timestamppb.New(event.To),
			Desc:       event.Description,
			OwnerId:    event.OwnerID,
			NotifyTime: timestamppb.New(event.Notify),
		}
	}
	return result
}
