package main

import (
	"context"
	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/thoas/go-funk"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/wantsToFress/first_stage_server/internal/model"
	event "github.com/wantsToFress/first_stage_server/pkg"
)

func modelToEvent(e *model.Event, groups []*model.Group, members []*model.Person) *event.Event {
	res := &event.Event{}

	res.Id = e.ID
	res.Name = stringToStringWrapper(e.Name)
	res.Description = ptrToStringWrapper(e.Description)
	res.Start = timeToTimestamp(e.Start)
	res.End = timeToTimestamp(e.End)
	res.CreatorId = stringToStringWrapper(e.CreatorID)

	if e.Creator != nil {
		res.Creator = modelToPersonListEntry(e.Creator)
	}

	res.Groups = make([]*event.GroupListEntry, 0, len(groups))
	res.Members = make([]*event.PersonListEntry, 0, len(members))

	for i := range groups {
		res.Groups = append(res.Groups, modelToGroupListEntry(groups[i]))
	}

	for i := range members {
		res.Members = append(res.Members, modelToPersonListEntry(members[i]))
	}

	return res
}

func modelToEventListEntry(e *model.Event) *event.EventListEntry {
	res := &event.EventListEntry{}

	res.Id = e.ID
	res.Name = stringToStringWrapper(e.Name)
	res.Description = ptrToStringWrapper(e.Description)
	res.Start = timeToTimestamp(e.Start)
	res.End = timeToTimestamp(e.End)
	res.CreatorId = stringToStringWrapper(e.CreatorID)

	return res
}

func (es *EventService) bindMembersToEvent(ctx context.Context, tx *pg.Tx, eventId string, memberIds []string) error {
	log := loggerFromContext(ctx)

	if len(memberIds) == 0 {
		return nil
	}

	groupMembers := make([]model.EventMember, 0, len(memberIds))
	for i := range memberIds {
		groupMembers = append(groupMembers, model.EventMember{
			PersonID: memberIds[i],
			EventID:  eventId,
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

func (es *EventService) unbindMembersToEvent(ctx context.Context, tx *pg.Tx, eventId string) error {
	log := loggerFromContext(ctx)

	_, err := tx.ModelContext(ctx, (*model.EventMember)(nil)).
		Where(model.Columns.EventMember.EventID+" = ?", eventId).
		Delete()
	if err != nil {
		log.WithError(err).Error("unable to delete group members")
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}

func (es *EventService) hasWriteAccessToEvent(ctx context.Context, eventId string) error {
	log := loggerFromContext(ctx)

	user, err := userFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("unable to get user from context")
		return err
	}

	e := &model.Event{}
	err = es.db.ModelContext(ctx, e).Where(model.Columns.Event.ID+" = ?", eventId).
		Select()
	if err != nil {
		log.WithError(err).Error("unable to select event")
		return status.Error(codes.Internal, err.Error())
	}

	if user.Role == event.Role_admin {
		return nil
	}

	if e.CreatorID == user.Id {
		return nil
	}

	if user.Role == event.Role_group_admin {
		err := es.db.ModelContext(ctx, &model.GroupEvent{}).
			Join("inner join "+model.Tables.GroupAdmin.Name+" as ga").
			JoinOn("t."+model.Columns.GroupEvent.GroupID+" = "+"ga."+model.Columns.GroupAdmin.GroupID).
			Where("t."+model.Columns.GroupEvent.EventID+" = ?", eventId).
			Where("ga."+model.Columns.GroupAdmin.PersonID+" = ?", user.Id).
			First()
		if err != nil {
			if err == pg.ErrNoRows {
				return status.Error(codes.PermissionDenied, "user has no write access to event")
			}
			return status.Error(codes.Internal, err.Error())
		}
		return nil
	}

	return status.Error(codes.PermissionDenied, "user has no write access to event")
}

func (es *EventService) hasReadAccessToEvent(ctx context.Context, eventId string) error {
	log := loggerFromContext(ctx)

	user, err := userFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("unable to get user from context")
		return err
	}

	if user.Role == event.Role_admin {
		return nil
	}

	e := &model.Event{}
	err = es.db.ModelContext(ctx, e).Where(model.Columns.Event.ID+" = ?", eventId).
		Select()
	if err != nil {
		log.WithError(err).Error("unable to select event")
		return status.Error(codes.Internal, err.Error())
	}

	if e.Type == event.EventType_OPENED.String() {
		return nil
	}

	if e.Type == event.EventType_GROUP.String() {
		switch user.Role {
		case event.Role_group_admin:
			{
				err := es.db.ModelContext(ctx, &model.GroupEvent{}).
					Join("inner join "+model.Tables.GroupAdmin.Name+" as ga").
					JoinOn("t."+model.Columns.GroupEvent.GroupID+" = "+"ga."+model.Columns.GroupAdmin.GroupID).
					Where("t."+model.Columns.GroupEvent.EventID+" = ?", eventId).
					Where("ga."+model.Columns.GroupAdmin.PersonID+" = ?", user.Id).
					First()
				if err != nil {
					if err != pg.ErrNoRows {
						return status.Error(codes.Internal, err.Error())
					}
				}
				err = es.db.ModelContext(ctx, &model.GroupEvent{}).
					Join("inner join "+model.Tables.GroupMember.Name+" as gm").
					JoinOn("t."+model.Columns.GroupEvent.GroupID+" = "+"gm."+model.Columns.GroupMember.GroupID).
					Where("t."+model.Columns.GroupEvent.EventID+" = ?", eventId).
					Where("gm."+model.Columns.GroupMember.PersonID+" = ?", user.Id).
					First()
				if err != nil {
					if err == pg.ErrNoRows {
						return status.Error(codes.PermissionDenied, "user has no read access to event")
					}
					return status.Error(codes.Internal, err.Error())
				}
			}
		case event.Role_group_member:
			{
				err := es.db.ModelContext(ctx, &model.GroupEvent{}).
					Join("inner join "+model.Tables.GroupMember.Name+" as gm").
					JoinOn("t."+model.Columns.GroupEvent.GroupID+" = "+"gm."+model.Columns.GroupMember.GroupID).
					Where("t."+model.Columns.GroupEvent.EventID+" = ?", eventId).
					Where("gm."+model.Columns.GroupMember.PersonID+" = ?", user.Id).
					First()
				if err != nil {
					if err == pg.ErrNoRows {
						return status.Error(codes.PermissionDenied, "user has no read access to event")
					}
					return status.Error(codes.Internal, err.Error())
				}
			}
		case event.Role_student:
			{
				return status.Error(codes.PermissionDenied, "user has no read access to event")
			}
		default:
			return status.Error(codes.PermissionDenied, "invalid role")
		}
		return nil
	}

	if e.Type == event.EventType_CLOSED.String() {
		err := es.db.ModelContext(ctx, &model.EventMember{}).
			Where(model.Columns.EventMember.EventID+" = ?", eventId).
			Where(model.Columns.EventMember.PersonID+" = ?", user.Id).
			First()
		if err != nil {
			if err == pg.ErrNoRows && e.CreatorID != user.Id {
				return status.Error(codes.PermissionDenied, "user has no read access to event")
			}
			return status.Error(codes.Internal, err.Error())
		}
		return nil
	}

	return status.Error(codes.PermissionDenied, "user has no read access to event")
}

func (es *EventService) CreateEvent(ctx context.Context, r *event.EventCreate) (*empty.Empty, error) {
	log := loggerFromContext(ctx)

	user, err := userFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("unable to get user from context")
		return nil, err
	}

	if r.GetStart() == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid start timestamp")
	}
	if r.GetEnd() == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid end timestamp")
	}

	e := &model.Event{}

	e.Name = r.GetName()
	e.Description = stringWrapperToPtr(r.GetDescription())
	e.Start = timestampToTime(r.GetStart())
	e.End = timestampToTime(r.GetEnd())
	e.CreatorID = user.Id
	e.Type = r.GetType().String()

	tx, err := es.db.WithContext(ctx).Begin()
	if err != nil {
		log.WithError(err).Error("unable to begin transaction")
		return nil, status.Error(codes.Internal, "unable to begin transaction")
	}

	_, err = tx.ModelContext(ctx, e).
		Insert()
	if err != nil {
		log.WithError(err).Error("unable to insert event")
		terr := tx.Rollback()
		if terr != nil {
			log.WithError(terr).Error("unable to rollback transaction")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = es.bindMembersToEvent(ctx, tx, e.ID, []string{e.CreatorID})
	if err != nil {
		terr := tx.Rollback()
		if terr != nil {
			log.WithError(terr).Error("unable to rollback transaction")
		}
		return nil, err
	}

	if len(r.GetMemberIds()) != 0 {
		err := es.bindMembersToEvent(ctx, tx, e.ID, r.GetMemberIds())
		if err != nil {
			terr := tx.Rollback()
			if terr != nil {
				log.WithError(terr).Error("unable to rollback transaction")
			}
			return nil, err
		}
	}

	if r.GetType() == event.EventType_GROUP {
		if !model.IsValidUUID(r.GetGroupId().GetValue()) {
			terr := tx.Rollback()
			if terr != nil {
				log.WithError(terr).Error("unable to rollback transaction")
			}
			return nil, status.Error(codes.InvalidArgument, "invalid group_id")
		}
		eventGroup := &model.GroupEvent{
			EventID: e.ID,
			GroupID: r.GetGroupId().GetValue(),
		}
		_, err := tx.ModelContext(ctx, eventGroup).
			OnConflict("do nothing").
			Insert()
		if err != nil {
			log.WithError(err).Error("unable to bind event to group")
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

func (es *EventService) UpdateEvent(ctx context.Context, r *event.EventUpdateRequest) (*empty.Empty, error) {
	log := loggerFromContext(ctx)

	if !model.IsValidUUID(r.GetEvent().GetId()) {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	err := es.hasWriteAccessToEvent(ctx, r.GetEvent().GetId())
	if err != nil {
		return nil, err
	}

	tx, err := es.db.WithContext(ctx).Begin()
	if err != nil {
		log.WithError(err).Error("unable to begin transaction")
		return nil, status.Error(codes.Internal, "unable to begin transaction")
	}

	query := tx.ModelContext(ctx, (*model.Event)(nil)).
		Where(model.Columns.Event.ID+" = ?", r.GetEvent().GetId())

	eventPrefix := "event."
	flag := false

	if funk.ContainsString(r.GetFieldMask().GetPaths(), eventPrefix+"name") {
		query.Set(model.Columns.Event.Name+" = ?", r.GetEvent().GetName().GetValue())
		flag = true
	}
	if funk.ContainsString(r.GetFieldMask().GetPaths(), eventPrefix+"description") {
		query.Set(model.Columns.Event.Description+" = ?", stringWrapperToPtr(r.GetEvent().GetDescription()))
		flag = true
	}
	if funk.ContainsString(r.GetFieldMask().GetPaths(), eventPrefix+"start") {
		query.Set(model.Columns.Event.Start+" = ?", timestampToTime(r.GetEvent().GetStart()))
		flag = true
	}
	if funk.ContainsString(r.GetFieldMask().GetPaths(), eventPrefix+"end") {
		query.Set(model.Columns.Event.End+" = ?", timestampToTime(r.GetEvent().GetEnd()))
		flag = true
	}

	if flag {
		_, err = query.Update()
		if err != nil {
			log.WithError(err).Error("unable to update event")
			terr := tx.Rollback()
			if terr != nil {
				log.WithError(terr).Error("unable to rollback transaction")
			}
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	if funk.ContainsString(r.GetFieldMask().GetPaths(), eventPrefix+"member_ids") {
		err := es.unbindMembersToEvent(ctx, tx, r.GetEvent().GetId())
		if err != nil {
			terr := tx.Rollback()
			if terr != nil {
				log.WithError(terr).Error("unable to rollback transaction")
			}
			return nil, err
		}
		err = es.bindMembersToEvent(ctx, tx, r.GetEvent().GetId(), r.GetEvent().GetMemberIds())
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

func (es *EventService) SetEventMembers(ctx context.Context, r *event.EventPersonsRequest) (*empty.Empty, error) {
	log := loggerFromContext(ctx)

	if !model.IsValidUUID(r.GetId()) {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	err := es.hasWriteAccessToEvent(ctx, r.GetId())
	if err != nil {
		return nil, err
	}

	tx, err := es.db.WithContext(ctx).Begin()
	if err != nil {
		log.WithError(err).Error("unable to begin transaction")
		return nil, status.Error(codes.Internal, "unable to begin transaction")
	}

	err = es.unbindMembersToEvent(ctx, tx, r.GetId())
	if err != nil {
		terr := tx.Rollback()
		if terr != nil {
			log.WithError(terr).Error("unable to rollback transaction")
		}
		return nil, err
	}
	err = es.bindMembersToEvent(ctx, tx, r.GetId(), r.GetPersonIds())
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

func (es *EventService) AddEventMembers(ctx context.Context, r *event.EventPersonsRequest) (*empty.Empty, error) {
	log := loggerFromContext(ctx)

	if !model.IsValidUUID(r.GetId()) {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	err := es.hasWriteAccessToEvent(ctx, r.GetId())
	if err != nil {
		return nil, err
	}

	tx, err := es.db.WithContext(ctx).Begin()
	if err != nil {
		log.WithError(err).Error("unable to begin transaction")
		return nil, status.Error(codes.Internal, "unable to begin transaction")
	}

	err = es.bindMembersToEvent(ctx, tx, r.GetId(), r.GetPersonIds())
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

func (es *EventService) GetEvent(ctx context.Context, r *event.Id) (*event.Event, error) {
	log := loggerFromContext(ctx)

	err := es.hasReadAccessToEvent(ctx, r.GetId())
	if err != nil {
		return nil, err
	}

	e := &model.Event{}
	err = es.db.ModelContext(ctx, e).
		Relation(model.Columns.Event.Creator).
		Where(model.Columns.Event.ID+" = ?", r.GetId()).
		Select()
	if err != nil {
		log.WithError(err).Error("unable to select event")
		return nil, status.Error(codes.Internal, err.Error())
	}

	groups := []*model.Group{}
	err = es.db.ModelContext(ctx, &groups).
		Distinct().
		ColumnExpr("t."+model.Columns.Group.ID).
		ColumnExpr("t."+model.Columns.Group.Name).
		ColumnExpr("t."+model.Columns.Group.Description).
		ColumnExpr("t."+model.Columns.Group.CreatedAt).
		ColumnExpr("t."+model.Columns.Group.UpdatedAt).
		Join("inner join "+model.Tables.GroupEvent.Name+" ad ge").
		JoinOn("t."+model.Columns.Group.ID+" = "+"ge."+model.Columns.GroupEvent.EventID).
		Where(model.Columns.GroupEvent.EventID+" = ?", r.GetId()).
		Select()
	if err != nil {
		log.WithError(err).Error("unable to select event groups")
		return nil, status.Error(codes.Internal, err.Error())
	}

	members := []*model.Person{}
	err = es.db.ModelContext(ctx, &members).
		Distinct().
		ColumnExpr("t."+model.Columns.Person.ID).
		ColumnExpr("t."+model.Columns.Person.FullName).
		ColumnExpr("t."+model.Columns.Person.Login).
		Join("inner join "+model.Tables.EventMember.Name+" ad ge").
		JoinOn("t."+model.Columns.Group.ID+" = "+"ge."+model.Columns.EventMember.EventID).
		Where(model.Columns.EventMember.EventID+" = ?", r.GetId()).
		Select()
	if err != nil {
		log.WithError(err).Error("unable to select event members")
		return nil, status.Error(codes.Internal, err.Error())
	}

	if e.Type == event.EventType_GROUP.String() && len(groups) != 0 {
		groupIds := make([]string, 0, len(groups))
		for i := range groups {
			groupIds = append(groupIds, groups[i].ID)
		}

		groupMembers := []*model.Person{}
		err = es.db.ModelContext(ctx, &groupMembers).
			Join("inner join "+model.Tables.GroupMember.Name+" as gm").
			JoinOn("t."+model.Columns.Person.ID+" = "+"gm."+model.Columns.GroupMember.PersonID).
			WhereIn("gm."+model.Columns.GroupMember.GroupID+" in(?)", groupIds).
			UnionAll(
				es.db.ModelContext(ctx, &groupMembers).
					Join("inner join "+model.Tables.GroupAdmin.Name+" as gm").
					JoinOn("t."+model.Columns.Person.ID+" = "+"gm."+model.Columns.GroupAdmin.PersonID).
					WhereIn("gm."+model.Columns.GroupAdmin.GroupID+" in(?)", groupIds),
			).Select()
		if err != nil {
			log.WithError(err).Error("unable to select event group members")
			return nil, status.Error(codes.Internal, err.Error())
		}
		members = append(members, groupMembers...)
	}

	return modelToEvent(e, groups, members), nil
}

func (es *EventService) ListEvents(ctx context.Context, r *event.EventListRequest) (*event.EventList, error) {
	log := loggerFromContext(ctx)

	user, err := userFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("unable to get user from context")
		return nil, err
	}

	events := []*model.Event{}
	query := es.db.ModelContext(ctx, &events).
		Distinct().
		ColumnExpr("t." + model.Columns.Event.ID).
		ColumnExpr("t." + model.Columns.Event.Name).
		ColumnExpr("t." + model.Columns.Event.Description).
		ColumnExpr("t." + model.Columns.Event.Start).
		ColumnExpr("t." + model.Columns.Event.End).
		ColumnExpr("t." + model.Columns.Event.Type).
		ColumnExpr("t." + model.Columns.Event.CreatorID)

	if r.GetName() != nil {
		query.Where(model.Columns.Event.Name+" ilike (?::text, '%')", r.GetName().GetValue())
	}

	switch user.Role {
	case event.Role_group_admin:
		query.Join("left join "+model.Tables.EventMember.Name+" as em").
			JoinOn("t."+model.Columns.Event.ID+" = "+"em."+model.Columns.EventMember.EventID).
			Join("left join "+model.Tables.GroupEvent.Name+" as ge").
			JoinOn("t."+model.Columns.Event.ID+" = "+"ge."+model.Columns.GroupEvent.EventID).
			Join("left join "+model.Tables.GroupMember.Name+" as gm").
			JoinOn("ge."+model.Columns.GroupEvent.GroupID+" is not null").
			JoinOn("ge."+model.Columns.GroupEvent.GroupID+" = "+"gm."+model.Columns.GroupMember.GroupID).
			Join("left join "+model.Tables.GroupAdmin.Name+" as ga").
			JoinOn("ge."+model.Columns.GroupEvent.GroupID+" is not null").
			JoinOn("ge."+model.Columns.GroupEvent.GroupID+" = "+"ga."+model.Columns.GroupAdmin.GroupID).
			WhereOr("t."+model.Columns.Event.Type+" = ?", event.EventType_OPENED.String()).
			WhereOrGroup(func(q *orm.Query) (*orm.Query, error) {
				return q.Where("em."+model.Columns.EventMember.PersonID+" is not null").
					Where("em."+model.Columns.EventMember.PersonID+" = ?", user.Id), nil
			}).
			WhereOrGroup(func(q *orm.Query) (*orm.Query, error) {
				return q.Where("gm."+model.Columns.GroupMember.PersonID+" is not null").
					Where("gm."+model.Columns.GroupMember.PersonID+" = ?", user.Id), nil
			}).
			WhereOrGroup(func(q *orm.Query) (*orm.Query, error) {
				return q.Where("ga."+model.Columns.GroupAdmin.PersonID+" is not null").
					Where("ga."+model.Columns.GroupAdmin.PersonID+" = ?", user.Id), nil
			})
	case event.Role_group_member:
		query.Join("left join "+model.Tables.EventMember.Name+" as em").
			JoinOn("t."+model.Columns.Event.ID+" = "+"em."+model.Columns.EventMember.EventID).
			Join("left join "+model.Tables.GroupEvent.Name+" as ge").
			JoinOn("t."+model.Columns.Event.ID+" = "+"ge."+model.Columns.GroupEvent.EventID).
			Join("left join "+model.Tables.GroupMember.Name+" as gm").
			JoinOn("ge."+model.Columns.GroupEvent.GroupID+" is not null").
			JoinOn("ge."+model.Columns.GroupEvent.GroupID+" = "+"gm."+model.Columns.GroupMember.GroupID).
			WhereOr("t."+model.Columns.Event.Type+" = ?", event.EventType_OPENED.String()).
			WhereOrGroup(func(q *orm.Query) (*orm.Query, error) {
				return q.Where("em."+model.Columns.EventMember.PersonID+" is not null").
					Where("em."+model.Columns.EventMember.PersonID+" = ?", user.Id), nil
			}).
			WhereOrGroup(func(q *orm.Query) (*orm.Query, error) {
				return q.Where("gm."+model.Columns.GroupMember.PersonID+" is not null").
					Where("gm."+model.Columns.GroupMember.PersonID+" = ?", user.Id), nil
			})
	case event.Role_student:
		query.Join("left join "+model.Tables.EventMember.Name+" as em").
			JoinOn("t."+model.Columns.Event.ID+" = "+"em."+model.Columns.EventMember.EventID).
			WhereOr("t."+model.Columns.Event.Type+" = ?", event.EventType_OPENED.String()).
			WhereOrGroup(func(q *orm.Query) (*orm.Query, error) {
				return q.Where("em."+model.Columns.EventMember.PersonID+" is not null").
					Where("em."+model.Columns.EventMember.PersonID+" = ?", user.Id), nil
			})
	default:
		return nil, status.Error(codes.PermissionDenied, "invalid role")
	}

	query, err = paginatedQuery(query, r.GetPagination(),
		model.Columns.Event.Name,
		model.Columns.Event.Start,
		model.Columns.Event.End,
		model.Columns.Event.CreatedAt,
		model.Columns.Event.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	totalHits, err := query.SelectAndCount()
	if err != nil {
		log.WithError(err).Error("unable to select events")
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := &event.EventList{
		Pagination: paginationWithHits(r.GetPagination(), totalHits),
		Events:     make([]*event.EventListEntry, 0, len(events)),
	}

	for i := range events {
		res.Events = append(res.Events, modelToEventListEntry(events[i]))
	}

	return res, nil
}

func (es *EventService) GetEventMembers(ctx context.Context, r *event.Id) (*event.PersonList, error) {
	e, err := es.GetEvent(ctx, r)
	if err != nil {
		return nil, err
	}
	return &event.PersonList{
		Persons: e.GetMembers(),
	}, nil
}

func (es *EventService) GetEventAdmins(ctx context.Context, r *event.Id) (*event.PersonList, error) {
	log := loggerFromContext(ctx)

	err := es.hasReadAccessToEvent(ctx, r.GetId())
	if err != nil {
		return nil, err
	}

	e := &model.Event{}
	err = es.db.ModelContext(ctx, e).
		Relation(model.Columns.Event.Creator).
		Where(model.Columns.Event.ID+" = ?", r.GetId()).
		Select()
	if err != nil {
		log.WithError(err).Error("unable to select event")
		return nil, status.Error(codes.Internal, err.Error())
	}

	if e.Type != event.EventType_GROUP.String() {
		person := &model.Person{}
		err := es.db.ModelContext(ctx, person).
			Where(model.Columns.Person.ID+" = ?", e.CreatorID).
			Select()
		if err != nil {
			log.WithError(err).Error("unable to select person")
			return nil, status.Error(codes.Internal, err.Error())
		}
		return &event.PersonList{
			Persons: []*event.PersonListEntry{
				modelToPersonListEntry(person),
			},
		}, nil
	}

	persons := []*model.Person{}

	err = es.db.ModelContext(ctx, persons).
		Distinct().
		ColumnExpr("t."+model.Columns.Person.ID).
		ColumnExpr("t."+model.Columns.Person.FullName).
		ColumnExpr("t."+model.Columns.Person.Login).
		ColumnExpr("t."+model.Columns.Person.CreatedAt).
		ColumnExpr("t."+model.Columns.Person.UpdatedAt).
		Join("inner join "+model.Tables.GroupAdmin.Name+" as ga").
		JoinOn("t."+model.Columns.Person.ID+" = "+"ga."+model.Columns.GroupAdmin.PersonID).
		Join("inner join "+model.Tables.GroupEvent.Name+" as ge").
		JoinOn("ga."+model.Columns.GroupAdmin.GroupID+" = "+"ge."+model.Columns.GroupEvent.GroupID).
		Where("ge."+model.Columns.GroupEvent.EventID+" = ?", r.GetId()).
		Select()
	if err != nil {
		log.WithError(err).Error("unable to select persons")
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := &event.PersonList{
		Persons: make([]*event.PersonListEntry, 0, len(persons)),
	}

	for i := range persons {
		res.Persons = append(res.Persons, modelToPersonListEntry(persons[i]))
	}

	return res, nil
}
