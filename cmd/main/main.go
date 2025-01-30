package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	handlersOfAnimal "pet_adopter/src/animal/handlers"
	logicOfAnimal "pet_adopter/src/animal/logic"
	repoOfAnimal "pet_adopter/src/animal/repo"
	handlersOfBreed "pet_adopter/src/breed/handlers"
	logicOfBreed "pet_adopter/src/breed/logic"
	repoOfBreed "pet_adopter/src/breed/repo"
	"pet_adopter/src/config"
	handlersOfLocality "pet_adopter/src/locality/handlers"
	logicOfLocality "pet_adopter/src/locality/logic"
	repoOfLocality "pet_adopter/src/locality/repo"
	"pet_adopter/src/middleware"
	handlersOfRegion "pet_adopter/src/region/handlers"
	logicOfRegion "pet_adopter/src/region/logic"
	repoOfRegion "pet_adopter/src/region/repo"
	handlersOfUser "pet_adopter/src/user/handlers"
	logicOfUser "pet_adopter/src/user/logic"
	repoOfUser "pet_adopter/src/user/repo"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Printf("failed to load .env file: %v", err)
	}
}

func main() {
	logFile, err := os.OpenFile(os.Getenv("MAIN_LOG_FILE"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("failed to open log file: %v", err)
		return
	}
	defer logFile.Close()

	logger := slog.New(slog.NewJSONHandler(io.MultiWriter(logFile, os.Stdout), &slog.HandlerOptions{Level: slog.LevelInfo}))
	logger.Info("Log file opened")

	cfg := config.MustLoadConfig(os.Getenv("CONFIG_FILE"), logger)
	logger.Info("Config file loaded")

	postgres, err := pgxpool.Connect(context.Background(), os.Getenv("POSTGRES_URL"))
	if err != nil {
		logger.Error(errors.Wrap(err, "failed to connect to postgres").Error())
		return
	}
	defer postgres.Close()

	if err = postgres.Ping(context.Background()); err != nil {
		logger.Error(errors.Wrap(err, "failed to ping postgres").Error())
		return
	}
	logger.Info("Postgres connected")

	redisOpts, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		logger.Error(errors.Wrap(err, "failed to parse redis url").Error())
		return
	}
	redisClient := redis.NewClient(redisOpts)

	if err = redisClient.Ping(context.Background()).Err(); err != nil {
		logger.Error(errors.Wrap(err, "failed to ping redis").Error())
		return
	}
	logger.Info("Redis connected")

	animalRepo := repoOfAnimal.NewAnimalPostgres(postgres)
	animalLogic := logicOfAnimal.NewAnimalLogic(animalRepo)
	animalHandler := handlersOfAnimal.NewAnimalHandler(animalLogic)

	breedRepo := repoOfBreed.NewBreedPostgres(postgres)
	breedLogic := logicOfBreed.NewBreedLogic(breedRepo)
	breedHandler := handlersOfBreed.NewBreedHandler(breedLogic)

	regionRepo := repoOfRegion.NewRegionPostgres(postgres)
	regionLogic := logicOfRegion.NewRegionLogic(regionRepo)
	regionHandler := handlersOfRegion.NewRegionHandler(regionLogic)

	localityRepo := repoOfLocality.NewLocalityPostgres(postgres)
	localityLogic := logicOfLocality.NewLocalityLogic(localityRepo)
	localityHandler := handlersOfLocality.NewLocalityHandler(localityLogic)

	sessionRepo := repoOfUser.NewSessionRedis(redisClient)
	sessionLogic := logicOfUser.NewSessionLogic(sessionRepo, cfg.Session)

	userRepo := repoOfUser.NewUserPostgres(postgres)
	userLogic := logicOfUser.NewUserLogic(userRepo)
	userHandler := handlersOfUser.NewUserHandler(userLogic, sessionLogic, cfg.Session, cfg.Validation)

	reqIDMiddleware := middleware.CreateRequestIDMiddleware(logger)
	sessionMiddleware := middleware.CreateSessionMiddleware(userLogic, sessionLogic, cfg.Session)

	r := mux.NewRouter().PathPrefix("/api/v1").Subrouter()
	r.Use(reqIDMiddleware, middleware.CorsMiddleware, middleware.RecoverMiddleware)

	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	user := r.PathPrefix("/user").Subrouter()
	{
		user.Handle("/signup", http.HandlerFunc(userHandler.SignUp)).
			Methods(http.MethodPost, http.MethodOptions)
		user.Handle("/login", http.HandlerFunc(userHandler.Login)).
			Methods(http.MethodPost, http.MethodOptions)
		user.Handle("/logout", sessionMiddleware(http.HandlerFunc(userHandler.Logout))).
			Methods(http.MethodPost, http.MethodOptions)
		user.Handle("", sessionMiddleware(http.HandlerFunc(userHandler.GetUser))).
			Methods(http.MethodGet, http.MethodOptions)
		user.Handle("/set_locality", sessionMiddleware(http.HandlerFunc(userHandler.SetLocality))).
			Methods(http.MethodPost, http.MethodOptions)
	}

	animals := r.PathPrefix("/animals").Subrouter()
	{
		animals.Handle("", http.HandlerFunc(animalHandler.GetAnimals)).
			Methods(http.MethodGet, http.MethodOptions)
		animals.Handle("/{id}", http.HandlerFunc(animalHandler.GetAnimalByID)).
			Methods(http.MethodGet, http.MethodOptions)
		animals.Handle("/add", middleware.AdminMiddleware(http.HandlerFunc(animalHandler.AddAnimal))).
			Methods(http.MethodPost, http.MethodOptions)
		animals.Handle("/remove", middleware.AdminMiddleware(http.HandlerFunc(animalHandler.RemoveAnimalByID))).
			Methods(http.MethodPost, http.MethodOptions)
	}

	breeds := r.PathPrefix("/breeds").Subrouter()
	{
		breeds.Handle("", http.HandlerFunc(breedHandler.GetBreeds)).
			Methods(http.MethodGet, http.MethodOptions)
		breeds.Handle("/{id}", http.HandlerFunc(breedHandler.GetBreedByID)).
			Methods(http.MethodGet, http.MethodOptions)
		breeds.Handle("/add", middleware.AdminMiddleware(http.HandlerFunc(breedHandler.AddBreed))).
			Methods(http.MethodPost, http.MethodOptions)
		breeds.Handle("/remove", middleware.AdminMiddleware(http.HandlerFunc(breedHandler.RemoveBreedByID))).
			Methods(http.MethodPost, http.MethodOptions)
	}

	regions := r.PathPrefix("/regions").Subrouter()
	{
		regions.Handle("", http.HandlerFunc(regionHandler.GetRegions)).
			Methods(http.MethodGet, http.MethodOptions)
		regions.Handle("/{id}", http.HandlerFunc(regionHandler.GetRegionByID)).
			Methods(http.MethodGet, http.MethodOptions)
		regions.Handle("/add", middleware.AdminMiddleware(http.HandlerFunc(regionHandler.AddRegion))).
			Methods(http.MethodPost, http.MethodOptions)
		regions.Handle("/remove", middleware.AdminMiddleware(http.HandlerFunc(regionHandler.RemoveRegionByID))).
			Methods(http.MethodPost, http.MethodOptions)
	}

	localities := r.PathPrefix("/localities").Subrouter()
	{
		localities.Handle("", http.HandlerFunc(localityHandler.GetLocalities)).
			Methods(http.MethodGet, http.MethodOptions)
		localities.Handle("/{id}", http.HandlerFunc(localityHandler.GetLocalityByID)).
			Methods(http.MethodGet, http.MethodOptions)
		localities.Handle("/add", middleware.AdminMiddleware(http.HandlerFunc(localityHandler.AddLocality))).
			Methods(http.MethodPost, http.MethodOptions)
		localities.Handle("/remove", middleware.AdminMiddleware(http.HandlerFunc(localityHandler.RemoveLocalityByID))).
			Methods(http.MethodPost, http.MethodOptions)
	}

	http.Handle("/", r)
	server := http.Server{
		Handler:           middleware.PathMiddleware(r),
		Addr:              fmt.Sprintf(":%s", cfg.Main.Port),
		ReadTimeout:       cfg.Main.ReadTimeout,
		WriteTimeout:      cfg.Main.WriteTimeout,
		ReadHeaderTimeout: cfg.Main.ReadHeaderTimeout,
		IdleTimeout:       cfg.Main.IdleTimeout,
	}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err = server.ListenAndServe(); err != nil {
			logger.Info("Server stopped")
		}
	}()
	logger.Info("Server started")

	sig := <-signalCh
	logger.Info("Received signal: " + sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Main.ShutdownTimeout)
	defer cancel()

	if err = server.Shutdown(ctx); err != nil {
		logger.Error(errors.Wrap(err, "failed to gracefully shutdown").Error())
	}
}
