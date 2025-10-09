package handlers

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Kost0/L3/internal/repository"
	"github.com/google/uuid"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

type Handler struct {
	DB *dbpg.DB
}

type GetURL struct {
	URL string `json:"url"`
}

func generateShortURL(longURL string) string {
	data := fmt.Sprintf("%s%d", longURL, time.Now().UnixNano())
	hash := md5.Sum([]byte(data))
	shortURL := hex.EncodeToString(hash[:])[:8]
	return shortURL
}

func (h *Handler) URLShortening(c *ginext.Context) {
	data, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	longURL := &GetURL{}

	err = json.Unmarshal(data, longURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	shortURL := generateShortURL(longURL.URL)

	linkUUID := uuid.New()

	zlog.Logger.Info().Msgf("Link UUID: %s", linkUUID)

	url := &repository.URL{
		UUID:     &linkUUID,
		ShortURL: shortURL,
		URL:      longURL.URL,
	}

	err = repository.InsertURL(h.DB, url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	zlog.Logger.Info().Msg("URL shortened")

	c.JSON(http.StatusOK, ginext.H{"shortURL": shortURL})
}

func (h *Handler) GoShortURL(c *ginext.Context) {
	shortURL := c.Param("short_url")

	zlog.Logger.Info().Msgf("Go Short URL: %s", shortURL)

	linkID, url, err := repository.GetLongURL(h.DB, shortURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	zlog.Logger.Info().Msg(linkID.String())

	ip := c.ClientIP()

	userAgent := c.GetHeader("User-Agent")

	infoUUID := uuid.New()

	info := &repository.URLInfo{
		UUID:      &infoUUID,
		LinkID:    &linkID,
		Time:      time.Now(),
		UserAgent: userAgent,
		IP:        ip,
	}

	err = repository.SaveInfo(h.DB, info)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	zlog.Logger.Info().Msg("Redirected to long url")

	c.Redirect(http.StatusFound, url)
}

func (h *Handler) Analytics(c *ginext.Context) {
	shortURL := c.Param("short_url")

	groupBy := c.Param("group")

	var param string

	switch groupBy {
	case "day":
		param = "DATE(time)"
	case "month":
		param = "DATE_TRUNC('month', time)"
	case "user_agent":
		param = "user_agent"
	default:
		param = ""
	}

	linkID, url, err := repository.GetLongURL(h.DB, shortURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	if param != "" {
		urlGroupsInfo := make([]repository.AnalyticsGroups, 0)

		urlGroupsInfo, err = repository.GetInfoWithGroup(h.DB, shortURL, param)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
			return
		}

		zlog.Logger.Info().Msg("URL information got")

		c.JSON(http.StatusOK, ginext.H{
			"id":      linkID,
			"longURL": url,
			"urlInfo": urlGroupsInfo,
		})
	} else {
		urlInfo := make([]repository.URLInfo, 0)

		urlInfo, err = repository.GetInfo(h.DB, shortURL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
			return
		}

		zlog.Logger.Info().Msg("URL information got")

		c.JSON(http.StatusOK, ginext.H{
			"id":      linkID,
			"longURL": url,
			"urlInfo": urlInfo,
		})
	}
}
