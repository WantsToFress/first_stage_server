package main

import (
	"context"
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
