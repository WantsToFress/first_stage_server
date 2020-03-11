package main

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/wantsToFress/first_stage_server/internal/model"
	event "github.com/wantsToFress/first_stage_server/pkg"
)

func modelToPersonListEntry(person *model.Person) *event.PersonListEntry {
	res := &event.PersonListEntry{}

	res.Id = person.ID
	res.Login = stringToStringWrapper(person.Login)
	res.FullName = stringToStringWrapper(person.FullName)

	if person.Role != nil {
		res.Role = event.Role(event.Role_value[person.Role.Name])
	}

	return res
}

func (es *EventService) GetPerson(context.Context, *event.Id) (*event.Person, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (es *EventService) getPersonByLogin(ctx context.Context, login string) (*event.PersonListEntry, error) {
	log := loggerFromContext(ctx)

	user := &model.Person{}
	err := es.db.ModelContext(ctx, user).
		Relation(model.Columns.Person.Role).
		Where(model.Columns.Person.Login+" = ?", login).
		Select()
	if err != nil {
		log.WithError(err).Error("unable to get person by login")
		return nil, status.Error(codes.Internal, err.Error())
	}

	return modelToPersonListEntry(user), nil
}

func (es *EventService) ListPerson(context.Context, *event.PersonListRequest) (*event.Person, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (es *EventService) WhoAmI(ctx context.Context, r *empty.Empty) (*event.Person, error) {
	user, err := userFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return es.GetPerson(ctx, &event.Id{Id: user.Id})
}

func (es *EventService) JoinEvent(context.Context, *event.PersonEventAssignment) (*empty.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (es *EventService) SetAdminGroups(context.Context, *event.PersonGroupRequest) (*empty.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (es *EventService) SetPersonRole(context.Context, *event.PersonRoleRequest) (*empty.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}
