package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/Kost0/L3/internal/middleware"
	"github.com/Kost0/L3/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/ginext"
)

type Handler struct {
	DB *dbpg.DB
}

func (h *Handler) CreateItem(c *ginext.Context) {
	role, exists := c.Get("role_id")
	if !exists {
		c.JSON(http.StatusInternalServerError, "role is not found")
		return
	}

	roleID, ok := role.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, "role is not string")
		return
	}

	data, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	newItemDTO := repository.ItemDTO{}

	err = json.Unmarshal(data, &newItemDTO)
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	newItem := repository.Item{
		UUID:      uuid.New(),
		Title:     newItemDTO.Title,
		Price:     newItemDTO.Price,
		Category:  newItemDTO.Category,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = repository.CreateItem(h.DB, &newItem, roleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, newItem)
}

func (h *Handler) GetItems(c *ginext.Context) {
	items, err := repository.GetAllItems(h.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, items)
}

func (h *Handler) UpdateItem(c *ginext.Context) {
	role, exists := c.Get("role_id")
	if !exists {
		c.JSON(http.StatusInternalServerError, "role is not found")
		return
	}

	roleID, ok := role.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, "role is not string")
		return
	}

	id := c.Param("id")

	data, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	updatedItemDTO := repository.ItemDTO{}

	err = json.Unmarshal(data, &updatedItemDTO)
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	updatedOrder := repository.ItemDTO{
		Title:    updatedItemDTO.Title,
		Price:    updatedItemDTO.Price,
		Category: updatedItemDTO.Category,
	}

	err = repository.UpdateItem(h.DB, &updatedOrder, id, roleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedOrder)
}

func (h *Handler) DeleteItem(c *ginext.Context) {
	role, exists := c.Get("role_id")
	if !exists {
		c.JSON(http.StatusInternalServerError, "role is not found")
		return
	}

	roleID, ok := role.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, "role is not string")
		return
	}

	id := c.Param("id")

	err := repository.DeleteItem(h.DB, id, roleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ginext.H{"order deleted": id})
}

func (h *Handler) Login(c *ginext.Context) {
	var loginRequest struct {
		Role string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
	}

	validRoles := map[string]bool{"admin": true, "manager": true, "viewer": true}
	if !validRoles[loginRequest.Role] {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "invalid role"})
		return
	}

	claims := middleware.Claims{
		Role: loginRequest.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func (h *Handler) GetHistory(c *ginext.Context) {
	id := c.Param("id")

	history, err := repository.GetHistory(h.DB, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
}
