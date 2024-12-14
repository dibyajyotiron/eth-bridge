package routers

import (
	"github.com/eth-bridging/internal/handlers"
	"github.com/eth-bridging/pkg/di"
	"github.com/gin-gonic/gin"
)

func SetupRouter(container *di.Container) *gin.Engine {
	router := gin.Default()

	eventHandler := handlers.NewBridgeEventHandler(container.EventService)

	apiV1 := router.Group("/api/v1")
	{
		apiV1.GET("/events", eventHandler.GetEvents)
	}

	return router
}
