package main

import (
	"context"
	"encoding/base64"
	"flag"
	"github.com/dgrijalva/jwt-go"
	stdlog "log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/go-chi/chi"
	"github.com/golang-migrate/migrate/v4"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/sirupsen/logrus"
	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"

	event "github.com/wantsToFress/first_stage_server/pkg"
)

const (
	configPath = "config.yaml"
)

func main() {
	configFile := flag.String("c", configPath, "specify path to a config.yaml")
	flag.Parse()

	logger := logrus.New()
	logger.Formatter = &logrus.JSONFormatter{}

	log := logrus.NewEntry(logger)

	config, err := Configure(*configFile)
	if err != nil {
		log.WithError(err).Fatal()
	}

	ctx := contextWithLogger(context.Background(), log)

	log.WithField("sas", config.Migration).Info()

	// migrate database
	err = Migrate(config.DB, config.Migration)
	if err != nil {
		if err != migrate.ErrNoChange && err != migrate.ErrNilVersion {
			log.WithError(err).Fatal("error on migrate")
		} else {
			log.Info("no actual migrations found")
		}
	} else {
		log.Info("all migrations was executed correctly")
	}

	// create db connection
	db, err := NewDBServer(ctx, config.DB)
	if err != nil {
		log.WithError(err).Fatal("unable to connect to db")
	}
	defer db.Finalize(ctx)

	// create centrifugo client
	cent, err := NewCentrifugoClient(config.Centrifuge)
	if err != nil {
		log.WithError(err).Fatal("unable to create centrifuge client")
	}
	defer cent.Close()

	pubkeyRaw, err := base64.StdEncoding.DecodeString(config.Auth.PublicKey)
	if err != nil {
		log.WithError(err).Fatal("unable to decode pubkey")
	}
	privkeyRaw, err := base64.StdEncoding.DecodeString(config.Auth.PrivateKey)
	if err != nil {
		log.WithError(err).Fatal("unable to decode privkey")
	}

	pubkey, err := jwt.ParseRSAPublicKeyFromPEM(pubkeyRaw)
	if err != nil {
		log.WithError(err).Fatal("unable to parse pubkey")
	}
	privkey, err := jwt.ParseRSAPrivateKeyFromPEM(privkeyRaw)
	if err != nil {
		log.WithError(err).Fatal("unable to parse privkey")
	}

	// back service
	ps := EventService{
		db:         db.Conn,
		cent:       cent,
		hmacSecret: []byte(config.Centrifuge.HMACSecret),
		privateKey: privkey,
		publicKey:  pubkey,
	}

	// grpc
	lis, err := net.Listen("tcp", config.Server.GrpcAddress)
	if err != nil {
		log.WithError(err).Fatal("failed to listen")
	}

	// person grpc server
	unaryInterceptors := []grpc.UnaryServerInterceptor{}
	unaryInterceptors = append(unaryInterceptors,
		ps.AuthInterceptor,
	)

	serverOptions := []grpc.ServerOption{}
	serverOptions = append(serverOptions, grpc.StatsHandler(&ocgrpc.ServerHandler{}))
	serverOptions = append(serverOptions, grpcmiddleware.WithUnaryServerChain(unaryInterceptors...))

	server := grpc.NewServer(serverOptions...)
	event.RegisterEventServiceServer(server, &ps)

	exit := make(chan error, 1)

	go func() {
		log.WithField("address", config.Server.GrpcAddress).Infof("Starting grpc server")
		err := server.Serve(lis)
		if err != nil {
			log.Error(err)
			exit <- err
		}
	}()
	defer server.GracefulStop()

	// grpc gateway

	gw := runtime.NewServeMux()

	conn, err := grpc.Dial(config.Server.GrpcAddress, grpc.WithInsecure())
	if err != nil {
		log.WithError(err).Fatal("unable to dial grpc.Client")
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			log.WithError(err).Fatal("unable to close grpc.Client connection")
		}
	}()
	client := event.NewEventServiceClient(conn)

	err = event.RegisterEventServiceHandlerClient(ctx, gw, client)
	if err != nil {
		log.Fatal(err)
	}

	router := chi.NewRouter()

	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("mycookie")
			if err == nil {
				r.Header.Set("Authorization", cookie.Value)
			}
			next.ServeHTTP(w, r)
		})
	})

	router.Post(path.Join(config.Server.BasePath, "login"), ps.Login)
	router.Post(path.Join(config.Server.BasePath, "register"), ps.Register)

	router.Mount(config.Server.BasePath, http.StripPrefix(config.Server.BasePath, gw))

	swaggerFullPath := path.Join(config.Server.BasePath, config.Swagger.Url)
	swaggerFullPathPrefix := swaggerFullPath + "/"
	fs := http.FileServer(http.Dir(config.Swagger.Path))
	router.Mount(swaggerFullPathPrefix, http.StripPrefix(swaggerFullPathPrefix, fs))
	router.Get(swaggerFullPath, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, swaggerFullPathPrefix, http.StatusMovedPermanently)
	})

	gatewayServer := &http.Server{
		Addr:     config.Server.GatewayAddress,
		Handler:  router,
		ErrorLog: stdlog.New(log.WithField("actor", "http.Server").Writer(), "", 0),
	}

	defer func() {
		err := gatewayServer.Shutdown(ctx)
		if err != nil {
			log.WithError(err).Warning("gateway shutdown failed")
		}
	}()
	go func() {
		log.WithField("address", config.Server.GatewayAddress).Info("Starting gateway http server")
		err := gatewayServer.ListenAndServe()
		if err != nil {
			log.Error(err)
			exit <- err
		}
	}()

	sgnl := make(chan os.Signal, 1)
	signal.Notify(sgnl,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	stop := <-sgnl
	log.Info("Received ", stop)
	log.Info("Waiting for stop all jobs")
}
