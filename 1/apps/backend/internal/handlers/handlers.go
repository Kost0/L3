package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/redis"
	"github.com/wb-go/wbf/retry"
	"internal/repository"
)

type Handler struct {
	Publisher   *rabbitmq.Publisher
	Manager     *rabbitmq.QueueManager
	DB          *dbpg.DB
	RedisClient *redis.Client
}

var retryStrategy = retry.Strategy{
	Attempts: 3,
	Delay:    time.Second,
	Backoff:  1.5,
}

func (h *Handler) GetNotify(c *ginext.Context) {
	id := c.Param("id")

	status, err := h.RedisClient.GetWithRetry(context.Background(), retryStrategy, id)
	if err != nil {
		if !errors.Is(err, redis.NoMatches) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		} else {
			status, err = repository.GetNotifyByID(id, h.DB)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"status": http.StatusText(http.StatusNotFound)})
				return
			}
		}
	}

	c.JSON(http.StatusOK, ginext.H{"notify": status})
}

func (h *Handler) CreateNotify(c *ginext.Context) {
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	newNotify := repository.Notify{}
	err = c.BindJSON(&newNotify)
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	if ok := isValidEmail(newNotify.Email); !ok {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "Invalid email"})
		return
	}

	timeSendAt, err := time.Parse(time.RFC3339, newNotify.SendAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "Invalid time"})
		return
	}

	err = repository.CreateNotify(&newNotify, h.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	err = h.RedisClient.SetWithRetry(context.Background(), retryStrategy, newNotify.ID, newNotify.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	delay := time.Until(timeSendAt)
	ttl := int32(delay)

	queueName := fmt.Sprintf("delay_%d", ttl)

	args := amqp091.Table{
		"x-message-ttl":             ttl,
		"x-dead-letter-exchange":    "notification-exchange",
		"x-dead-letter-routing-key": "notify-key",
		"x-expires":                 ttl + 60000,
	}

	queueConfig := rabbitmq.QueueConfig{
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args:       args,
	}

	_, err = h.Manager.DeclareQueue(queueName, queueConfig)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	err = h.Publisher.PublishWithRetry(body, queueName, "application/json", retryStrategy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ginext.H{"notify": body})
}

func (h *Handler) DeleteNotify(c *ginext.Context) {
	id := c.Param("id")

	err := repository.DeleteNotifyByID(id, h.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, ginext.H{"result": "notify deleted"})
}
