package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/eth-bridging/internal/services"

	"github.com/gin-gonic/gin"
)

type BridgeEventHandler struct {
	service services.BridgeEventService
}

func NewBridgeEventHandler(service services.BridgeEventService) *BridgeEventHandler {
	return &BridgeEventHandler{
		service: service,
	}
}

func (h *BridgeEventHandler) GetEvents(c *gin.Context) {
	// Get query parameters
	lastIDStr := c.Query("last_id")
	limitStr := c.Query("limit")

	// Default limit if not provided
	limit := 10
	// Max limit if high value is provided
	maxLimit := 100

	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
			return
		}
		fmt.Println(parsedLimit, maxLimit)
		if parsedLimit > maxLimit {
			parsedLimit = maxLimit
		}
		limit = parsedLimit
	}

	// Convert lastID to uint
	var lastID uint
	if lastIDStr != "" {
		parsedID, err := strconv.ParseUint(lastIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid last_id parameter"})
			return
		}
		lastID = uint(parsedID)
	}

	// Fetch events
	events, err := h.service.GetAllEvents(lastID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(events) > 0 {
		lastEvent := events[len(events)-1]
		lastID = uint(lastEvent.ID)
	}

	// Return the response with pagination info
	c.JSON(http.StatusOK, gin.H{
		"events":  events,
		"last_id": lastID,
	})
}
