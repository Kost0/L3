package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Kost0/L3/internal/repository"
	"github.com/google/uuid"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

type Handler struct {
	DB *dbpg.DB
}

type GetComment struct {
	Text   string     `json:"text"`
	Parent *uuid.UUID `json:"parent"`
}

func (h *Handler) CreateComment(c *ginext.Context) {
	data, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	getComment := &GetComment{}

	err = json.Unmarshal(data, getComment)
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	commentUUID := uuid.New()

	comment := &repository.Comment{
		UUID:   &commentUUID,
		Text:   getComment.Text,
		Parent: getComment.Parent,
	}

	err = repository.InsertComment(h.DB, comment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ginext.H{"Comment": comment.UUID})
}

func (h *Handler) GetComments(c *ginext.Context) {
	parent := c.DefaultQuery("parent", "null")

	comments, err := repository.SelectComments(h.DB, parent)
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, comments)
}

func (h *Handler) GetPageComments(c *ginext.Context) {
	zlog.Logger.Info().Msg("Getting page of comments...")

	page, err := strconv.Atoi(c.DefaultQuery("page", "0"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	sort := c.DefaultQuery("sort", "")

	comments, err := repository.SelectPage(h.DB, page, sort)
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	pages, err := repository.CountPages(h.DB)

	c.JSON(http.StatusOK, ginext.H{
		"comments":   comments,
		"totalPages": pages,
	})
}

func (h *Handler) DeleteComment(c *ginext.Context) {
	id := c.Param("id")

	err := repository.DeleteComments(h.DB, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ginext.H{"deleted": id})
}

func (h *Handler) SearchComment(c *ginext.Context) {
	keyword := c.DefaultQuery("query", "null")

	comments, err := repository.SearchComments(h.DB, keyword)
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, comments)
}
