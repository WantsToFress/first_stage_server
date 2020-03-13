package main

import (
	"context"
	"encoding/json"
	"github.com/centrifugal/centrifuge-go"
	"github.com/prometheus/common/log"
	"time"

	"github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/wantsToFress/first_stage_server/internal/model"
	event "github.com/wantsToFress/first_stage_server/pkg"
)

func (es *EventService) GetChatToken(ctx context.Context, r *event.Id) (*event.ChatToken, error) {
	log := loggerFromContext(ctx)

	if !model.IsValidUUID(r.GetId()) {
		return nil, status.Error(codes.InvalidArgument, "invalid id")
	}

	err := es.hasReadAccessToEvent(ctx, r.GetId())
	if err != nil {
		return nil, err
	}

	user, err := userFromContext(ctx)
	if err != nil {
		log.WithError(err).Error("unable to get user from context")
		return nil, err
	}

	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = jwt.MapClaims{
		"sub": user.Login,
		"exp": time.Now().Add(time.Hour * 24),
	}
	tokenRaw, err := token.SignedString(es.hmacSecret)
	if err != nil {
		return nil, err
	}

	return &event.ChatToken{Token: tokenRaw}, nil
}

type Message struct {
	Id       string `json:"id"`
	FullName string `json:"full_name"`
	UID      string `json:"uid"`
	Login    string `json:"login"`
	EventId  string `json:"event_id"`
	Time     int64  `json:"time"`
	Message  string `json:"message"`
}

func modelToMessage(m *model.Message) *event.Message {
	res := &event.Message{}

	res.Id = m.ID
	res.Message = m.Message
	res.Time = m.Time.Unix()
	res.Login = m.Login
	res.EventId = m.EventID
	res.FullName = m.FullName
	res.Uid = m.PersonID

	return res
}

func (es *EventService) GetChatHistory(ctx context.Context, r *event.Id) (*event.ChatHistory, error) {
	log := loggerFromContext(ctx)

	messages := []*model.Message{}
	err := es.db.ModelContext(ctx, &messages).
		Where(model.Columns.Message.EventID+" = ?", r.GetId()).
		Order(model.Columns.Message.Time + " ASC").
		Select()
	if err != nil {
		log.WithError(err).Error("unable to select chat")
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := &event.ChatHistory{
		Messages: make([]*event.Message, 0, len(messages)),
	}

	for i := range messages {
		res.Messages = append(res.Messages, modelToMessage(messages[i]))
	}

	return res, nil
}

func (es *EventService) OnPublish(sub *centrifuge.Subscription, e centrifuge.PublishEvent) {
	data, err := e.Data.MarshalJSON()
	if err != nil {
		log.Error(err)
		return
	}
	msg := &Message{}
	err = json.Unmarshal(data, msg)
	if err != nil {
		log.Error(msg)
		return
	}

	message := &model.Message{
		ID:       msg.Id,
		PersonID: msg.UID,
		EventID:  msg.EventId,
		Login:    msg.Login,
		FullName: msg.FullName,
		Time:     time.Unix(msg.Time/1000, 0),
		Message:  msg.Message,
	}

	_, err = es.db.Model(message).
		OnConflict("do nothing").
		Insert()
	if err != nil {
		log.Error(err)
	}
}

func (es *EventService) WatchChat(ctx context.Context) error {
	sub, err := es.cent.NewSubscription("all")
	if err != nil {
		return err
	}

	sub.OnPublish(es)

	err = sub.Subscribe()
	if err != nil {
		return nil
	}

	return nil
}
