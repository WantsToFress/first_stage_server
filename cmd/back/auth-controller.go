package main

import (
	"context"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-pg/pg/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/wantsToFress/first_stage_server/internal/model"
	event "github.com/wantsToFress/first_stage_server/pkg"
)

type tokenContextKey struct{}

func userToContext(ctx context.Context, user *event.PersonListEntry) context.Context {
	return context.WithValue(ctx, tokenContextKey{}, user)
}

func userFromContext(ctx context.Context) (*event.PersonListEntry, error) {
	user, ok := ctx.Value(tokenContextKey{}).(*event.PersonListEntry)
	if !ok {
		return nil, status.Error(codes.Internal, "user not present in context")
	}
	return user, nil
}

type UserClaims struct {
	Login      string    `json:"login"`
	Expiration time.Time `json:"exp"`
}

func (u *UserClaims) Valid() error {
	if u.Login == "" {
		return status.Error(codes.InvalidArgument, "empty login")
	}
	return nil
}

func (es *EventService) AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata not present in context")
	}

	authMeta := meta.Get("authorization")
	if len(authMeta) == 0 {
		return nil, status.Error(codes.Unauthenticated, "authorization header not present")
	}

	authRaw := authMeta[0]

	token, err := jwt.ParseWithClaims(authRaw, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return es.publicKey, nil
	})
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	userClaims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, status.Error(codes.Internal, "invalid claims type")
	}

	user, err := es.getPersonByLogin(ctx, userClaims.Login)
	if err != nil {
		return nil, err
	}

	return handler(userToContext(ctx, user), req)
}

func (es *EventService) newToken(login string) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodRS256, &UserClaims{
		Login:      login,
		Expiration: time.Now().Add(time.Hour * 24 * 3),
	}).SignedString(es.privateKey)
}

func (es *EventService) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	login := r.FormValue("login")
	full_name := r.FormValue("full_name")
	password := r.FormValue("password")

	role := &model.Role{}
	err := es.db.ModelContext(ctx, role).
		Where(model.Columns.Role.Name+" = ?", event.Role_student.String()).
		Select()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user := &model.Person{
		FullName: full_name,
		Login:    login,
		Password: password,
		RoleID:   role.ID,
	}

	_, err = es.db.ModelContext(ctx, user).
		Insert()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := es.newToken(login)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", token)
	w.WriteHeader(http.StatusOK)
}

func (es *EventService) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	login := r.FormValue("login")
	password := r.FormValue("password")

	err := es.db.ModelContext(ctx, &model.Person{}).
		Where(model.Columns.Person.Login+" = ?", login).
		Where(model.Columns.Person.Password+" = ?", password).
		First()
	if err != nil {
		if err == pg.ErrNoRows {
			http.Error(w, "user with given credentials doesn't exist", http.StatusUnauthorized)
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	token, err := es.newToken(login)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", token)
	w.WriteHeader(http.StatusOK)
}
