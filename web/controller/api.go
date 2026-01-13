package controller

import (
	"net/http"

	"github.com/mhsanaei/3x-ui/v2/web/service"
	"github.com/mhsanaei/3x-ui/v2/web/session"

	"github.com/gin-gonic/gin"
)

type TesterBearer interface {
	TestBearer(bearer string) bool
}

// APIController handles the main API routes for the 3x-ui panel, including inbounds and server management.
type APIController struct {
	BaseController
	bearer            TesterBearer
	inboundController *InboundController
	serverController  *ServerController
	Tgbot             service.Tgbot
}

// NewAPIController creates a new APIController instance and initializes its routes.
func NewAPIController(b TesterBearer, g *gin.RouterGroup) *APIController {
	a := &APIController{bearer: b}
	a.initRouter(g)
	return a
}

// checkAPIAuth is a middleware that returns 404 for unauthenticated API requests
// to hide the existence of API endpoints from unauthorized users
func (a *APIController) checkAPIAuth(c *gin.Context) {
	if a.bearer.TestBearer(c.GetHeader("Authorization")) || session.IsLogin(c) {
		c.Next()
		return
	}

	c.AbortWithStatus(http.StatusNotFound)
}

// initRouter sets up the API routes for inbounds, server, and other endpoints.
func (a *APIController) initRouter(g *gin.RouterGroup) {
	// Main API group
	api := g.Group("/panel/api")
	api.Use(a.checkAPIAuth)

	// Inbounds API
	inbounds := api.Group("/inbounds")
	a.inboundController = NewInboundController(inbounds)

	// Server API
	server := api.Group("/server")
	a.serverController = NewServerController(server)

	// Extra routes
	api.GET("/backuptotgbot", a.BackuptoTgbot)
}

// BackuptoTgbot sends a backup of the panel data to Telegram bot admins.
func (a *APIController) BackuptoTgbot(c *gin.Context) {
	a.Tgbot.SendBackupToAdmins()
}
