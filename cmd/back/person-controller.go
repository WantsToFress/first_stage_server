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

func modelToPerson(person *model.Person, memberGroups, adminGroups []*model.Group) *event.Person {
	res := &event.Person{}

	res.Id = person.ID
	res.Login = stringToStringWrapper(person.Login)
	res.FullName = stringToStringWrapper(person.FullName)
	res.StudentGroups = make([]*event.GroupListEntry, 0, len(memberGroups))
	res.AdminGroups = make([]*event.GroupListEntry, 0, len(adminGroups))

	if person.Role != nil {
		res.Role = event.Role(event.Role_value[person.Role.Name])
	}

	for i := range memberGroups {
		res.StudentGroups = append(res.StudentGroups, modelToGroupListEntry(memberGroups[i]))
	}

	for i := range adminGroups {
		res.AdminGroups = append(res.AdminGroups, modelToGroupListEntry(adminGroups[i]))
	}

	return res
}

func (es *EventService) GetPerson(ctx context.Context, r *event.Id) (*event.Person, error) {
	log := loggerFromContext(ctx)

	if !model.IsValidUUID(r.GetId()) {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	user := &model.Person{}
	err := es.db.ModelContext(ctx, user).
		Relation(model.Columns.Person.Role).
		Where(model.Columns.Person.ID+" = ?", r.GetId()).
		Select()
	if err != nil {
		log.WithError(err).Error("unable to get person by login")
		return nil, status.Error(codes.Internal, err.Error())
	}

	memberGroups := []*model.Group{}

	err = es.db.ModelContext(ctx, memberGroups).
		Distinct().
		ColumnExpr("t."+model.Columns.Group.ID).
		ColumnExpr("t."+model.Columns.Group.Name).
		ColumnExpr("t."+model.Columns.Group.Decsription).
		Join("inner join "+model.Tables.GroupMember.Name+" as gm").
		JoinOn("t."+model.Columns.Group.ID+" = "+"gm."+model.Columns.GroupMember.GroupID).
		Where(model.Columns.GroupMember.PersonID+" = ?", r.GetId()).
		Select()
	if err != nil {
		log.WithError(err).Error("unable to select member groups")
		return nil, status.Error(codes.Internal, err.Error())
	}

	adminGroups := []*model.Group{}

	err = es.db.ModelContext(ctx, adminGroups).
		Distinct().
		ColumnExpr("t."+model.Columns.Group.ID).
		ColumnExpr("t."+model.Columns.Group.Name).
		ColumnExpr("t."+model.Columns.Group.Decsription).
		Join("inner join "+model.Tables.GroupAdmin.Name+" as gm").
		JoinOn("t."+model.Columns.Group.ID+" = "+"gm."+model.Columns.GroupAdmin.GroupID).
		Where(model.Columns.GroupAdmin.PersonID+" = ?", r.GetId()).
		Select()
	if err != nil {
		log.WithError(err).Error("unable to select member groups")
		return nil, status.Error(codes.Internal, err.Error())
	}

	return modelToPerson(user, memberGroups, adminGroups), nil
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

func (es *EventService) ListPersons(ctx context.Context, r *event.PersonListRequest) (*event.PersonList, error) {
	log := loggerFromContext(ctx)

	users := []*model.Person{}

	query := es.db.ModelContext(ctx, &users)

	if r.GetSearch() != nil {
		query.WhereOr(model.Columns.Person.FullName+" ilike concat(?::text, '%')", r.GetSearch().GetValue()).
			WhereOr(model.Columns.Person.Login+" ilike concat(?::text, '%')", r.GetSearch().GetValue())
	}

	totalHits, err := query.SelectAndCount()
	if err != nil {
		log.WithError(err).Error("unable to select persons")
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := &event.PersonList{
		Pagination: paginationWithHits(r.GetPagination(), totalHits),
		Persons:    make([]*event.PersonListEntry, 0, len(users)),
	}
	for i := range users {
		res.Persons = append(res.Persons, modelToPersonListEntry(users[i]))
	}

	return res, nil
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

func (es *EventService) SetAdminGroups(ctx context.Context, r *event.PersonGroupRequest) (*empty.Empty, error) {
	log := loggerFromContext(ctx)

	if !model.IsValidUUID(r.GetId()) {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	currentGroups := []string{}
	err := es.db.ModelContext(ctx, (*model.GroupAdmin)(nil)).
		ColumnExpr(model.Columns.GroupAdmin.GroupID).
		Where(model.Columns.GroupAdmin.PersonID+" = ?", r.GetId()).
		Select(&currentGroups)
	if err != nil {
		log.WithError(err).Error("unable to get current person groups")
		return nil, status.Error(codes.Internal, err.Error())
	}

	for i := range currentGroups {
		err := es.hasWriteAccessToGroup(ctx, currentGroups[i])
		if err != nil {
			return nil, err
		}
	}

	for i := range r.GetGroupIds() {
		err := es.hasWriteAccessToGroup(ctx, r.GetGroupIds()[i])
		if err != nil {
			return nil, err
		}
	}

	tx, err := es.db.WithContext(ctx).Begin()
	if err != nil {
		log.WithError(err).Error("unable to begin transaction")
		return nil, status.Error(codes.Internal, "unable to begin transaction")
	}

	_, err = tx.ModelContext(ctx, (*model.GroupAdmin)(nil)).
		Where(model.Columns.GroupAdmin.PersonID+" = ?", r.GetId()).
		Delete()
	if err != nil {
		log.WithError(err).Error("unable to delete old person admin groups")
		terr := tx.Rollback()
		if terr != nil {
			log.WithError(terr).Error("unable to rollback transaction")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if len(r.GetGroupIds()) != 0 {
		newGroups := make([]model.GroupAdmin, 0, len(r.GetGroupIds()))
		for i := range r.GetGroupIds() {
			newGroups = append(newGroups, model.GroupAdmin{
				PersonID: r.GetId(),
				GroupID:  r.GetGroupIds()[i],
			})
		}
		_, err := tx.ModelContext(ctx, &newGroups).
			OnConflict("do nothing").
			Insert()
		if err != nil {
			log.WithError(err).Error("unable to insert new admin groups")
			terr := tx.Rollback()
			if terr != nil {
				log.WithError(terr).Error("unable to rollback transaction")
			}
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	err = tx.Commit()
	if err != nil {
		log.WithError(err).Error("unable to commit transaction")
		return nil, status.Error(codes.Internal, "unable to commit transaction")
	}

	return &empty.Empty{}, nil
}

func (es *EventService) SetPersonRole(ctx context.Context, r *event.PersonRoleRequest) (*empty.Empty, error) {
	log := loggerFromContext(ctx)

	if !model.IsValidUUID(r.GetId()) {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	user, err := userFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("unable to get user from context")
		return nil, err
	}

	switch r.GetRole() {
	case event.Role_admin, event.Role_group_admin:
		{
			if user.Role != event.Role_admin {
				return nil, status.Error(codes.PermissionDenied, "user can not assign specified role")
			}
		}
	case event.Role_group_member:
		{
			if user.Role != event.Role_admin && user.Role != event.Role_group_admin {
				return nil, status.Error(codes.PermissionDenied, "user can not assign specified role")
			}
		}
	}

	tx, err := es.db.WithContext(ctx).Begin()
	if err != nil {
		log.WithError(err).Error("unable to begin transaction")
		return nil, status.Error(codes.Internal, "unable to begin transaction")
	}

	role := &model.Role{}
	err = tx.ModelContext(ctx, role).
		Where(model.Columns.Role.Name+" = ?", r.GetRole().String()).
		Select()
	if err != nil {
		log.WithError(err).Error("unable to select role")
		terr := tx.Rollback()
		if terr != nil {
			log.WithError(terr).Error("unable to rollback transaction")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	_, err = tx.ModelContext(ctx, &model.Person{}).
		Where(model.Columns.Person.ID+" = ?", r.GetId()).
		Set(model.Columns.Person.RoleID+" = ?", role.ID).
		Update()
	if err != nil {
		log.WithError(err).Error("unable to update person role")
		terr := tx.Rollback()
		if terr != nil {
			log.WithError(terr).Error("unable to rollback transaction")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = tx.Commit()
	if err != nil {
		log.WithError(err).Error("unable to commit transaction")
		return nil, status.Error(codes.Internal, "unable to commit transaction")
	}

	return &empty.Empty{}, nil
}
