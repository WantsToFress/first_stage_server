package main

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	event "github.com/wantsToFress/first_stage_server/pkg"
)

func (es *EventService) CreateEvent(context.Context, *event.EventCreate) (*empty.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (es *EventService) UpdateEvent(context.Context, *event.EventUpdateRequest) (*empty.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (es *EventService) SetEventMembers(context.Context, *event.EventPersonsRequest) (*empty.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (es *EventService) GetEvent(context.Context, *event.Id) (*event.Event, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (es *EventService) ListEvents(context.Context, *event.EventListRequest) (*event.EventList, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (es *EventService) GetEventMembers(context.Context, *event.Id) (*event.PersonList, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (es *EventService) GetEventAdmins(context.Context, *event.Id) (*event.PersonList, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}
