package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Kost0/L3/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/redis"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
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
	zlog.Logger.Info().Msg("Getting notify status...")

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

	zlog.Logger.Info().Msg("Status got")

	c.JSON(http.StatusOK, ginext.H{"notify": status})
}

func (h *Handler) CreateNotify(c *ginext.Context) {
	zlog.Logger.Info().Msg("Creating notify...")

	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	zlog.Logger.Info().Msg("Read data")

	newNotify := repository.Notify{}
	err = json.Unmarshal(body, &newNotify)
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	newNotify.Status = "waiting"

	newNotify.ID = uuid.New().String()

	zlog.Logger.Info().Msg("Unmarshalled")

	if ok := isValidEmail(newNotify.Email); !ok {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "Invalid email"})
		return
	}

	layout := "2006-01-02T15:04"
	//         2025-10-03T17:42

	timeSendAt, err := time.Parse(layout, newNotify.SendAt)
	if err != nil || timeSendAt.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "Invalid time"})
		return
	}

	err = repository.CreateNotify(&newNotify, h.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	zlog.Logger.Info().Msg("Db saved data")

	err = h.RedisClient.SetWithRetry(context.Background(), retryStrategy, newNotify.ID, newNotify.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	zlog.Logger.Info().Msg("Redis saved data")

	delay := timeSendAt.Sub(time.Now())
	ttl := delay.Milliseconds()

	queueName := fmt.Sprintf("delay_%d", ttl)

	args := amqp091.Table{
		"x-message-ttl":             ttl,
		"x-dead-letter-exchange":    "notification-exchange",
		"x-dead-letter-routing-key": "notify-key",
		"x-expires":                 ttl + 60000,
	}

	zlog.Logger.Info().Msg(fmt.Sprintf("%d", delay))

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

	zlog.Logger.Info().Msg("Queue created")

	err = h.Publisher.PublishWithRetry(body, queueName, "application/json", retryStrategy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	zlog.Logger.Info().Msg("Data published")

	zlog.Logger.Info().Msg("Notify created, id = " + newNotify.ID)

	c.JSON(http.StatusCreated, ginext.H{"notify": body})
}

func (h *Handler) DeleteNotify(c *ginext.Context) {
	zlog.Logger.Info().Msg("Deleting notify...")

	id := c.Param("id")

	err := repository.DeleteNotifyByID(id, h.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, ginext.H{"result": "notify deleted"})
}
