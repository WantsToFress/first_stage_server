package main

import (
	"context"

	"github.com/go-pg/pg/v9"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/thoas/go-funk"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/wantsToFress/first_stage_server/internal/model"
	event "github.com/wantsToFress/first_stage_server/pkg"
)

func (es *EventService) bindMembersToGroup(ctx context.Context, tx *pg.Tx, groupId string, memberIds []string) error {
	log := loggerFromContext(ctx)

	if len(memberIds) == 0 {
		return nil
	}

	groupMembers := make([]model.GroupMember, 0, len(memberIds))
	for i := range memberIds {
		groupMembers = append(groupMembers, model.GroupMember{
			PersonID: memberIds[i],
			GroupID:  groupId,
		})
	}

	_, err := tx.ModelContext(ctx, &groupMembers).
		OnConflict("do nothing").
		Insert()
	if err != nil {
		log.WithError(err).Error("unable to insert group members")
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}

func (es *EventService) unbindMembersToGroup(ctx context.Context, tx *pg.Tx, groupId string) error {
	log := loggerFromContext(ctx)

	_, err := tx.ModelContext(ctx, (*model.GroupMember)(nil)).
		Where(model.Columns.GroupMember.GroupID+" = ?", groupId).
		Delete()
	if err != nil {
		log.WithError(err).Error("unable to delete group members")
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}

func (es *EventService) bindAdminsToGroup(ctx context.Context, tx *pg.Tx, groupId string, memberIds []string) error {
	log := loggerFromContext(ctx)

	if len(memberIds) == 0 {
		return nil
	}

	groupMembers := make([]model.GroupAdmin, 0, len(memberIds))
	for i := range memberIds {
		groupMembers = append(groupMembers, model.GroupAdmin{
			PersonID: memberIds[i],
			GroupID:  groupId,
		})
	}

	_, err := tx.ModelContext(ctx, &groupMembers).
		OnConflict("do nothing").
		Insert()
	if err != nil {
		log.WithError(err).Error("unable to insert group admins")
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}

func (es *EventService) unbindAdminsToGroup(ctx context.Context, tx *pg.Tx, groupId string) error {
	log := loggerFromContext(ctx)

	_, err := tx.ModelContext(ctx, (*model.GroupAdmin)(nil)).
		Where(model.Columns.GroupAdmin.GroupID+" = ?", groupId).
		Delete()
	if err != nil {
		log.WithError(err).Error("unable to delete group admins")
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}

func (es *EventService) CreateGroup(ctx context.Context, r *event.GroupCreate) (*empty.Empty, error) {
	log := loggerFromContext(ctx)

	user, err := userFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("unable to get user from context")
		return nil, err
	}

	if user.Role != event.Role_admin {
		return nil, status.Error(codes.PermissionDenied, "only admins can create groups")
	}

	if r.GetName() == nil {
		return nil, status.Error(codes.InvalidArgument, "group name could not be null")
	}

	tx, err := es.db.WithContext(ctx).Begin()
	if err != nil {
		log.WithError(err).Error("unable to begin transaction")
		return nil, status.Error(codes.Internal, "unable to begin transaction")
	}

	group := &model.Group{
		Name:        r.GetName().GetValue(),
		Description: stringWrapperToPtr(r.GetDescription()),
	}

	_, err = tx.ModelContext(ctx, group).
		Insert()
	if err != nil {
		log.WithError(err).Error("unable to insert group")
		terr := tx.Rollback()
		if terr != nil {
			log.WithError(terr).Error("unable to rollback transaction")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	if len(r.GetMemberIds()) != 0 {
		err := es.bindMembersToGroup(ctx, tx, group.ID, r.GetMemberIds())
		if err != nil {
			log.WithError(err).Error("unable to insert group members")
			terr := tx.Rollback()
			if terr != nil {
				log.WithError(terr).Error("unable to rollback transaction")
			}
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		log.WithError(err).Error("unable to commit transaction")
		return nil, status.Error(codes.Internal, "unable to commit transaction")
	}

	return &empty.Empty{}, nil
}

func (es *EventService) hasWriteAccessToGroup(ctx context.Context, groupId string) error {
	log := loggerFromContext(ctx)

	user, err := userFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("unable to get user from context")
		return err
	}

	if user.Role == event.Role_admin {
		return nil
	}

	if user.Role == event.Role_group_admin {
		err := es.db.ModelContext(ctx, &model.GroupAdmin{}).
			Where(model.Columns.GroupAdmin.GroupID+" = ?", groupId).
			Where(model.Columns.GroupAdmin.PersonID+" = ?", user.Id).
			First()
		if err != nil {
			if err == pg.ErrNoRows {
				return status.Error(codes.PermissionDenied, "user has no write access to group")
			}
			return status.Error(codes.Internal, err.Error())
		}
	}

	return status.Error(codes.PermissionDenied, "user has no write access to group")
}

func (es *EventService) hasReadAccessToGroup(ctx context.Context, groupId string) error {
	log := loggerFromContext(ctx)

	user, err := userFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("unable to get user from context")
		return err
	}

	if user.Role == event.Role_admin {
		return nil
	}

	if user.Role == event.Role_group_admin {
		err := es.db.ModelContext(ctx, &model.GroupAdmin{}).
			Where(model.Columns.GroupAdmin.GroupID+" = ?", groupId).
			Where(model.Columns.GroupAdmin.PersonID+" = ?", user.Id).
			First()
		if err != nil {
			if err == pg.ErrNoRows {
				err := es.db.ModelContext(ctx, &model.GroupMember{}).
					Where(model.Columns.GroupMember.GroupID+" = ?", groupId).
					Where(model.Columns.GroupMember.PersonID+" = ?", user.Id).
					First()
				if err != nil {
					if err == pg.ErrNoRows {
						return status.Error(codes.PermissionDenied, "user has no read access to group")
					}
					log.WithError(err).Error("unable to determine group access")
					return status.Error(codes.Internal, err.Error())
				}
			} else {
				log.WithError(err).Error("unable to determine group access")
				return status.Error(codes.Internal, err.Error())
			}
		}
		return nil
	}

	if user.Role == event.Role_group_member {
		err := es.db.ModelContext(ctx, &model.GroupMember{}).
			Where(model.Columns.GroupMember.GroupID+" = ?", groupId).
			Where(model.Columns.GroupMember.PersonID+" = ?", user.Id).
			First()
		if err != nil {
			if err == pg.ErrNoRows {
				return status.Error(codes.PermissionDenied, "user has no read access to group")
			}
			log.WithError(err).Error("unable to determine group access")
			return status.Error(codes.Internal, err.Error())
		}
		return nil
	}

	return status.Error(codes.PermissionDenied, "user has no read access to group")
}

func (es *EventService) UpdateGroup(ctx context.Context, r *event.GroupUpdateRequest) (*empty.Empty, error) {
	log := loggerFromContext(ctx)

	if !model.IsValidUUID(r.GetGroup().GetId()) {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	err := es.hasWriteAccessToGroup(ctx, r.GetGroup().GetId())
	if err != nil {
		return nil, err
	}

	tx, err := es.db.WithContext(ctx).Begin()
	if err != nil {
		log.WithError(err).Error("unable to begin transaction")
		return nil, status.Error(codes.Internal, "unable to begin transaction")
	}

	query := tx.ModelContext(ctx, &model.Group{}).
		Where(model.Columns.Group.ID+" = ?", r.GetGroup().GetId())

	groupPrefix := "group."
	flag := false

	if funk.ContainsString(r.GetFieldMask().GetPaths(), groupPrefix+"name") {
		query.Set(model.Columns.Group.Name+" = ?", r.GetGroup().GetName())
		flag = true
	}

	if funk.ContainsString(r.GetFieldMask().GetPaths(), groupPrefix+"description") {
		query.Set(model.Columns.Group.Description+" = ?", stringWrapperToPtr(r.GetGroup().GetDescription()))
		flag = true
	}

	if flag {
		_, err = query.Update()
		if err != nil {
			log.WithError(err).Error("unable to update group")
			terr := tx.Rollback()
			if terr != nil {
				log.WithError(terr).Error("unable to rollback transaction")
			}
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	if funk.ContainsString(r.GetFieldMask().GetPaths(), groupPrefix+"member_ids") {
		err := es.unbindMembersToGroup(ctx, tx, r.GetGroup().GetId())
		if err != nil {
			terr := tx.Rollback()
			if terr != nil {
				log.WithError(terr).Error("unable to rollback transaction")
			}
			return nil, err
		}
		err = es.bindMembersToGroup(ctx, tx, r.GetGroup().GetId(), r.GetGroup().GetMemberIds())
		if err != nil {
			terr := tx.Rollback()
			if terr != nil {
				log.WithError(terr).Error("unable to rollback transaction")
			}
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		log.WithError(err).Error("unable to commit transaction")
		return nil, status.Error(codes.Internal, "unable to commit transaction")
	}

	return &empty.Empty{}, nil
}

func (es *EventService) SetGroupMembers(ctx context.Context, r *event.GroupPersonsRequest) (*empty.Empty, error) {
	log := loggerFromContext(ctx)

	if !model.IsValidUUID(r.GetId()) {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	err := es.hasWriteAccessToGroup(ctx, r.GetId())
	if err != nil {
		return nil, err
	}

	tx, err := es.db.WithContext(ctx).Begin()
	if err != nil {
		log.WithError(err).Error("unable to begin transaction")
		return nil, status.Error(codes.Internal, "unable to begin transaction")
	}

	err = es.unbindMembersToGroup(ctx, tx, r.GetId())
	if err != nil {
		terr := tx.Rollback()
		if terr != nil {
			log.WithError(terr).Error("unable to rollback transaction")
		}
		return nil, err
	}
	err = es.bindMembersToGroup(ctx, tx, r.GetId(), r.GetPersonIds())
	if err != nil {
		terr := tx.Rollback()
		if terr != nil {
			log.WithError(terr).Error("unable to rollback transaction")
		}
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		log.WithError(err).Error("unable to commit transaction")
		return nil, status.Error(codes.Internal, "unable to commit transaction")
	}

	return &empty.Empty{}, nil
}

func (es *EventService) SetGroupAdmins(ctx context.Context, r *event.GroupPersonsRequest) (*empty.Empty, error) {
	log := loggerFromContext(ctx)

	if !model.IsValidUUID(r.GetId()) {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	err := es.hasWriteAccessToGroup(ctx, r.GetId())
	if err != nil {
		return nil, err
	}

	tx, err := es.db.WithContext(ctx).Begin()
	if err != nil {
		log.WithError(err).Error("unable to begin transaction")
		return nil, status.Error(codes.Internal, "unable to begin transaction")
	}

	err = es.unbindAdminsToGroup(ctx, tx, r.GetId())
	if err != nil {
		terr := tx.Rollback()
		if terr != nil {
			log.WithError(terr).Error("unable to rollback transaction")
		}
		return nil, err
	}
	err = es.bindAdminsToGroup(ctx, tx, r.GetId(), r.GetPersonIds())
	if err != nil {
		terr := tx.Rollback()
		if terr != nil {
			log.WithError(terr).Error("unable to rollback transaction")
		}
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		log.WithError(err).Error("unable to commit transaction")
		return nil, status.Error(codes.Internal, "unable to commit transaction")
	}

	return &empty.Empty{}, nil
}

func (es *EventService) AddGroupAdmins(ctx context.Context, r *event.GroupPersonsRequest) (*empty.Empty, error) {
	log := loggerFromContext(ctx)

	if !model.IsValidUUID(r.GetId()) {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	err := es.hasWriteAccessToGroup(ctx, r.GetId())
	if err != nil {
		return nil, err
	}

	tx, err := es.db.WithContext(ctx).Begin()
	if err != nil {
		log.WithError(err).Error("unable to begin transaction")
		return nil, status.Error(codes.Internal, "unable to begin transaction")
	}

	err = es.bindAdminsToGroup(ctx, tx, r.GetId(), r.GetPersonIds())
	if err != nil {
		terr := tx.Rollback()
		if terr != nil {
			log.WithError(terr).Error("unable to rollback transaction")
		}
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		log.WithError(err).Error("unable to commit transaction")
		return nil, status.Error(codes.Internal, "unable to commit transaction")
	}

	return &empty.Empty{}, nil
}

func (es *EventService) AddGroupMembers(ctx context.Context, r *event.GroupPersonsRequest) (*empty.Empty, error) {
	log := loggerFromContext(ctx)

	if !model.IsValidUUID(r.GetId()) {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	err := es.hasWriteAccessToGroup(ctx, r.GetId())
	if err != nil {
		return nil, err
	}

	tx, err := es.db.WithContext(ctx).Begin()
	if err != nil {
		log.WithError(err).Error("unable to begin transaction")
		return nil, status.Error(codes.Internal, "unable to begin transaction")
	}

	err = es.unbindMembersToGroup(ctx, tx, r.GetId())
	if err != nil {
		terr := tx.Rollback()
		if terr != nil {
			log.WithError(terr).Error("unable to rollback transaction")
		}
		return nil, err
	}
	err = es.bindMembersToGroup(ctx, tx, r.GetId(), r.GetPersonIds())
	if err != nil {
		terr := tx.Rollback()
		if terr != nil {
			log.WithError(terr).Error("unable to rollback transaction")
		}
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		log.WithError(err).Error("unable to commit transaction")
		return nil, status.Error(codes.Internal, "unable to commit transaction")
	}

	return &empty.Empty{}, nil
}

func modelToGroup(group *model.Group, members []*event.PersonListEntry, admins []*event.PersonListEntry) *event.Group {
	res := &event.Group{}

	res.Id = group.ID
	res.Name = stringToStringWrapper(group.Name)
	res.Description = ptrToStringWrapper(group.Description)
	res.Members = members
	res.Admins = admins

	return res
}

func modelToGroupListEntry(group *model.Group) *event.GroupListEntry {
	res := &event.GroupListEntry{}

	res.Id = group.ID
	res.Name = stringToStringWrapper(group.Name)
	res.Description = ptrToStringWrapper(group.Description)

	return res
}

func (es *EventService) GetGroup(ctx context.Context, r *event.Id) (*event.Group, error) {
	log := loggerFromContext(ctx)

	err := es.hasReadAccessToGroup(ctx, r.GetId())
	if err != nil {
		return nil, err
	}

	group := &model.Group{}

	err = es.db.ModelContext(ctx, group).
		Where(model.Columns.Group.ID+" = ?", r.GetId()).
		Select()
	if err != nil {
		log.WithError(err).Error("unable to select group")
		return nil, status.Error(codes.Internal, err.Error())
	}

	members, err := es.GetGroupMembers(ctx, &event.Id{Id: r.GetId()})
	if err != nil {
		return nil, err
	}

	admins, err := es.GetGroupAdmins(ctx, &event.Id{Id: r.GetId()})
	if err != nil {
		return nil, err
	}

	return modelToGroup(group, members.GetPersons(), admins.GetPersons()), nil
}

func (es *EventService) ListGroups(ctx context.Context, r *event.GroupListRequest) (*event.GroupList, error) {
	log := loggerFromContext(ctx)

	groups := []*model.Group{}

	query := es.db.ModelContext(ctx, &groups)

	if r.GetName() != nil {
		query.Where(model.Columns.Group.Name+" ilike concat(?::text, '%')", r.GetName().GetValue())
	}

	query, err := paginatedQuery(query, r.GetPagination(),
		model.Columns.Group.Name,
		model.Columns.Group.CreatedAt,
		model.Columns.Group.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	totalHist, err := query.SelectAndCount()
	if err != nil {
		log.WithError(err).Error("unable to select persons")
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := &event.GroupList{
		Pagination: paginationWithHits(r.GetPagination(), totalHist),
		Groups:     make([]*event.GroupListEntry, 0, len(groups)),
	}
	for i := range groups {
		res.Groups = append(res.Groups, modelToGroupListEntry(groups[i]))
	}

	return res, nil
}

func (es *EventService) GetGroupMembers(ctx context.Context, r *event.Id) (*event.PersonList, error) {
	log := loggerFromContext(ctx)

	err := es.hasReadAccessToGroup(ctx, r.GetId())
	if err != nil {
		return nil, err
	}

	members := []*model.Person{}

	err = es.db.ModelContext(ctx, &members).
		Distinct().
		Join("inner join "+model.Tables.GroupMember.Name+" as p").
		JoinOn("p."+model.Columns.GroupMember.PersonID+" = "+"t."+model.Columns.Person.ID).
		ColumnExpr("t."+model.Columns.Person.ID).
		ColumnExpr("t."+model.Columns.Person.FullName).
		ColumnExpr("t."+model.Columns.Person.Login).
		Where("p."+model.Columns.GroupMember.GroupID+" = ?", r.GetId()).
		Select()
	if err != nil {
		log.WithError(err).Error("unable to select members")
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := &event.PersonList{
		Persons: make([]*event.PersonListEntry, 0, len(members)),
	}
	for i := range members {
		res.Persons = append(res.Persons, modelToPersonListEntry(members[i]))
	}

	return res, nil
}

func (es *EventService) GetGroupAdmins(ctx context.Context, r *event.Id) (*event.PersonList, error) {
	log := loggerFromContext(ctx)

	err := es.hasReadAccessToGroup(ctx, r.GetId())
	if err != nil {
		return nil, err
	}

	admins := []*model.Person{}

	err = es.db.ModelContext(ctx, &admins).
		Distinct().
		Join("inner join "+model.Tables.GroupAdmin.Name+" as p").
		JoinOn("p."+model.Columns.GroupAdmin.PersonID+" = "+"t."+model.Columns.Person.ID).
		ColumnExpr("t."+model.Columns.Person.ID).
		ColumnExpr("t."+model.Columns.Person.FullName).
		ColumnExpr("t."+model.Columns.Person.Login).
		Where("p."+model.Columns.GroupAdmin.GroupID+" = ?", r.GetId()).
		Select()
	if err != nil {
		log.WithError(err).Error("unable to select members")
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := &event.PersonList{
		Persons: make([]*event.PersonListEntry, 0, len(admins)),
	}
	for i := range admins {
		res.Persons = append(res.Persons, modelToPersonListEntry(admins[i]))
	}

	return res, nil
}
