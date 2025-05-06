package main

//coverage:ignore file
import (
	_ "backend/docs"
	"backend/internal/handler/bundlesHandler"
	"backend/internal/handler/loginHandler"
	"backend/internal/handler/logoutHandler"
	"backend/internal/handler/productsHandler"
	"backend/internal/handler/rankingsHandler"
	"backend/internal/handler/registrationHandler"
	"backend/internal/handler/testsHandler"
	"backend/internal/handler/userProfileHandler"
	"backend/internal/handler/usersHandler"
	"backend/internal/middleware"
	"backend/internal/services"
	"github.com/swaggo/http-swagger"
	"log"
	"net/http"
)

func main() {
	// Initialize DB
	db := services.InitDB()
	defer db.Close()

	// Initialize middleware
	auth := middleware.NewAuthHandler(db)
	logger := middleware.NewLoggingHandler(db)

	// Initialize handlers (utilizes dependency injection)
	tests := testsHandler.TestsHandler(db)
	products := productsHandler.ProductsHandler(db)
	rankings := rankingsHandler.RankingsHandler(db)
	registration := registrationHandler.RegistrationHandler(db)
	login := loginHandler.LoginHandler(db)
	logout := logoutHandler.LogoutHandler(db)
	bundles := bundlesHandler.BundlesHandler(db)
	users := usersHandler.UsersHandler(db)
	userProfile := userProfileHandler.UserProfileHandler(db)
	//session := http.HandlerFunc(sessionHandler.IsSessionActive)

	// Create a new ServeMux to handle routes.
	mux := http.NewServeMux()

	// Register endpoints.
	// Wrap each handler with the LoggingMiddleware.
	mux.Handle("/register", logger.LoggingMiddleware(registration))
	mux.Handle("/login", logger.LoggingMiddleware(login))
	mux.Handle("/logout", auth.Middleware(logger.LoggingMiddleware(logout)))
	mux.Handle("/tests", auth.Middleware(logger.LoggingMiddleware(tests)))
	mux.Handle("/tests/", auth.Middleware(logger.LoggingMiddleware(tests)))
	mux.Handle("/products", auth.Middleware(logger.LoggingMiddleware(products)))
	mux.Handle("/products/", auth.Middleware(logger.LoggingMiddleware(products)))
	mux.Handle("/rankings", auth.Middleware(logger.LoggingMiddleware(rankings)))
	mux.Handle("/rankings/", auth.Middleware(logger.LoggingMiddleware(rankings)))
	mux.Handle("/bundles", auth.Middleware(logger.LoggingMiddleware(bundles)))
	mux.Handle("/bundles/", auth.Middleware(logger.LoggingMiddleware(bundles)))
	mux.Handle("/users/", auth.Middleware(logger.LoggingMiddleware(users)))
	mux.Handle("/user/profile", auth.Middleware(logger.LoggingMiddleware(userProfile)))
	//mux.Handle("/is-session-active", auth.Middleware(session))

	// Swagger documentation route
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	// Start the HTTP server on port 8080.
	log.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
