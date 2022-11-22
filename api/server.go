package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	db "github.com/jyoonie/simplebank/db/sqlc"
	"github.com/jyoonie/simplebank/token"
	"github.com/jyoonie/simplebank/util"
)

// Server serves HTTP requests for our banking service.
type Server struct {
	config     util.Config
	store      db.Store // jy: will allow us to interact with the db whe processing api requests from clients
	tokenMaker token.Maker
	router     *gin.Engine // jy: will help us send each api request to the correct handler for processing
}

// NewServer creates a new HTTP server and setup routing.
func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	server.setupRouter()
	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()

	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)

	authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker))

	authRoutes.POST("/accounts", server.createAccount)
	authRoutes.GET("/accounts/:id", server.getAccount)
	authRoutes.GET("/accounts", server.ListAccounts)
	// authRoutes.PUT("/accounts/:id", server.updateAccount)
	// authRoutes.DELETE("/accounts/:id", server.deleteAccount)

	authRoutes.POST("/transfers", server.createTransfer)

	server.router = router

}

// Start runs the HTTP server on a specific address.
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
