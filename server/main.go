package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
	"github.com/momomo0206/go-chat-app/db"
	migration "github.com/momomo0206/go-chat-app/db/migrations"
	coreHandler "github.com/momomo0206/go-chat-app/internal/api/handler/core"
	statsHandler "github.com/momomo0206/go-chat-app/internal/api/handler/stats"
	userHandler "github.com/momomo0206/go-chat-app/internal/api/handler/user"
	roomRepo "github.com/momomo0206/go-chat-app/internal/repo/room"
	statsRepo "github.com/momomo0206/go-chat-app/internal/repo/stats"
	repository "github.com/momomo0206/go-chat-app/internal/repo/user"
	"github.com/momomo0206/go-chat-app/internal/service/pinnedrooms"
	statsService "github.com/momomo0206/go-chat-app/internal/service/stats"
	service "github.com/momomo0206/go-chat-app/internal/service/user"
	"github.com/momomo0206/go-chat-app/internal/ws"
	"github.com/momomo0206/go-chat-app/router"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variable")
	}

	dbConn, err := db.NewDatabase()
	if err != nil {
		log.Fatalf("Could not initialize DB connection: %s", err)
	}
	defer dbConn.Close()

	if err := dbConn.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to database successfully")

	// Run migrations
	if err := migration.RunMigrations(dbConn); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Set up Repositories
	userRepo := repository.NewUserRepository(dbConn)
	statsRepository := statsRepo.NewStatsRepository(dbConn)

	// Set up Services
	userService := service.NewUserService(userRepo)
	statsServ := statsService.NewStatsService(statsRepository)
	wsService := ws.NewCore(dbConn)

	// Set up Handlers
	userHandler := userHandler.NewUserHandler(userService)
	coreHandler := coreHandler.NewCoreHandler(wsService)
	statsHand := statsHandler.NewStatsHandler(statsServ)

	go wsService.Run()

	pinnedRoomsService := pinnedrooms.NewPinnedRoomsService(dbConn, wsService)
	if err := pinnedRoomsService.CheckAndRefreshPinnedRooms(context.Background()); err != nil {
		log.Printf("Failed to initialize pinned rooms: %v", err)
	}

	// Start background job to clean up expired rooms
	go startRoomCleanupJob(dbConn, wsService)

	router := router.SetupRouter(userHandler, coreHandler, statsHand)
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func startRoomCleanupJob(db *sql.DB, wsCore *ws.Core) {
	roomRepository := roomRepo.NewRoomRepository(db)
	pinnedRoomsService := pinnedrooms.NewPinnedRoomsService(db, wsCore)
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	cleanupRooms(roomRepository, pinnedRoomsService)

	for range ticker.C {
		cleanupRooms(roomRepository, pinnedRoomsService)
	}
}

func cleanupRooms(roomRepository *roomRepo.RoomRepository, pinnedRoomsService *pinnedrooms.PinnedRoomsService) {
	ctx := context.Background()
	deletedCount, err := roomRepository.DeleteExpiredRooms(ctx)
	if err != nil {
		log.Printf("Error deleting expired rooms: %v", err)
		return
	}

	if deletedCount > 0 {
		log.Printf("Deleted %d expired rooms", deletedCount)
	}

	if err := pinnedRoomsService.CheckAndRefreshPinnedRooms(ctx); err != nil {
		log.Printf("Error refreshing pinned rooms: %v", err)
	}
}
