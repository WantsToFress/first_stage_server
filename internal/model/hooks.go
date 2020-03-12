package model

import (
	"context"
	"github.com/go-pg/pg/v9/orm"
)

// Event hook
var _ orm.BeforeInsertHook = (*Event)(nil)
var _ orm.BeforeUpdateHook = (*Event)(nil)

func (model *Event) BeforeInsert(ctx context.Context) (context.Context, error) {
	model.ID = GenStringUUID()
	return ctx, nil
}

func (model *Event) BeforeUpdate(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

// Person hook
var _ orm.BeforeInsertHook = (*Person)(nil)
var _ orm.BeforeUpdateHook = (*Person)(nil)

func (model *Person) BeforeInsert(ctx context.Context) (context.Context, error) {
	model.ID = GenStringUUID()
	return ctx, nil
}

func (model *Person) BeforeUpdate(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

// Group hook
var _ orm.BeforeInsertHook = (*Group)(nil)
var _ orm.BeforeUpdateHook = (*Group)(nil)

func (model *Group) BeforeInsert(ctx context.Context) (context.Context, error) {
	model.ID = GenStringUUID()
	return ctx, nil
}

func (model *Group) BeforeUpdate(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

// GroupAdmin hook
var _ orm.BeforeInsertHook = (*GroupAdmin)(nil)
var _ orm.BeforeUpdateHook = (*GroupAdmin)(nil)

func (model *GroupAdmin) BeforeInsert(ctx context.Context) (context.Context, error) {
	model.ID = GenStringUUID()
	return ctx, nil
}

func (model *GroupAdmin) BeforeUpdate(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

// GroupEvent hook
var _ orm.BeforeInsertHook = (*GroupEvent)(nil)
var _ orm.BeforeUpdateHook = (*GroupEvent)(nil)

func (model *GroupEvent) BeforeInsert(ctx context.Context) (context.Context, error) {
	model.ID = GenStringUUID()
	return ctx, nil
}

func (model *GroupEvent) BeforeUpdate(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

// GroupMember hook
var _ orm.BeforeInsertHook = (*GroupMember)(nil)
var _ orm.BeforeUpdateHook = (*GroupMember)(nil)

func (model *GroupMember) BeforeInsert(ctx context.Context) (context.Context, error) {
	model.ID = GenStringUUID()
	return ctx, nil
}

func (model *GroupMember) BeforeUpdate(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

// Role hook
var _ orm.BeforeInsertHook = (*Role)(nil)
var _ orm.BeforeUpdateHook = (*Role)(nil)

func (model *Role) BeforeInsert(ctx context.Context) (context.Context, error) {
	model.ID = GenStringUUID()
	return ctx, nil
}

func (model *Role) BeforeUpdate(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

// Role hook
var _ orm.BeforeInsertHook = (*EventMember)(nil)
var _ orm.BeforeUpdateHook = (*EventMember)(nil)

func (model *EventMember) BeforeInsert(ctx context.Context) (context.Context, error) {
	model.ID = GenStringUUID()
	return ctx, nil
}

func (model *EventMember) BeforeUpdate(ctx context.Context) (context.Context, error) {
	return ctx, nil
}
