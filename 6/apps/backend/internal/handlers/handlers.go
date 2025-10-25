package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Kost0/L3/internal/repository"
	"github.com/google/uuid"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/ginext"
)

type Handler struct {
	DB *dbpg.DB
}

func (h *Handler) CreateOrder(c *ginext.Context) {
	data, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	newOrderDTO := repository.OrderDTO{}

	err = json.Unmarshal(data, &newOrderDTO)
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	newOrder := repository.Order{
		OrderID:  uuid.New(),
		Title:    newOrderDTO.Title,
		Cost:     newOrderDTO.Cost,
		Items:    newOrderDTO.Items,
		Category: newOrderDTO.Category,
		Date:     newOrderDTO.Date,
	}

	err = repository.CreateOrder(h.DB, &newOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, newOrder)
}

func (h *Handler) GetOrders(c *ginext.Context) {
	orders, err := repository.GetAllOrders(h.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)
}

func (h *Handler) UpdateOrder(c *ginext.Context) {
	id := c.Param("id")

	orderUUID, err := uuid.FromBytes([]byte(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	data, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	updatedOrderDTO := repository.OrderDTO{}

	err = json.Unmarshal(data, &updatedOrderDTO)
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	updatedOrder := repository.Order{
		OrderID:  orderUUID,
		Title:    updatedOrderDTO.Title,
		Cost:     updatedOrderDTO.Cost,
		Items:    updatedOrderDTO.Items,
		Category: updatedOrderDTO.Category,
		Date:     updatedOrderDTO.Date,
	}

	err = repository.UpdateOrder(h.DB, &updatedOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedOrder)
}

func (h *Handler) DeleteOrder(c *ginext.Context) {
	id := c.Param("id")

	err := repository.DeleteOrder(h.DB, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ginext.H{"order deleted": id})
}

func (h *Handler) GetAnalytics(c *ginext.Context) {
	from := c.DefaultQuery("from", time.Now().Add(-24*time.Hour).Format(time.RFC3339))
	to := c.DefaultQuery("to", time.Now().Add(24*time.Hour).Format(time.RFC3339))

	sumCost, err := repository.GetSum(h.DB, "cost", from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	sumItems, err := repository.GetSum(h.DB, "items", from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	avgCost, err := repository.GetAvg(h.DB, "cost", from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	avgItems, err := repository.GetAvg(h.DB, "items", from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	count, err := repository.GetCount(h.DB, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	medianCost, err := repository.GetMedian(h.DB, "cost", from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	medianItems, err := repository.GetMedian(h.DB, "items", from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	perCost, err := repository.GetPercentile(h.DB, "cost", from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	perItems, err := repository.GetPercentile(h.DB, "items", from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ginext.H{
		"count":        count,
		"sum_cost":     sumCost,
		"avg_cost":     avgCost,
		"median_cost":  medianCost,
		"per_cost":     perCost,
		"sum_items":    sumItems,
		"avg_items":    avgItems,
		"median_items": medianItems,
		"per_items":    perItems,
	})
}

func (h *Handler) GetAllCategories(c *ginext.Context) {
	categories, err := repository.GetAllCategories(h.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, categories)
}

func (h *Handler) GetOrdersByCategory(c *ginext.Context) {
	category := c.Param("category")

	orders, err := repository.GetOrdersByCategory(h.DB, category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)
}
