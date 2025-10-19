package handlers

import (
	"io"
	"net/http"

	"github.com/Kost0/L3/internal/minIO"
	"github.com/Kost0/L3/internal/repository"
	"github.com/Kost0/L3/internal/startKafka"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/segmentio/kafka-go"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/ginext"
)

type Handler struct {
	DB     *dbpg.DB
	Client *minio.Client
	Writer *kafka.Writer
}

func (h *Handler) ProcessPhoto(c *ginext.Context) {
	fileHeader, err := c.FormFile("imageFile")
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	defer file.Close()

	imageUUID := uuid.New()
	bucketName := "images"
	objectKey := imageUUID.String()
	contentType := fileHeader.Header.Get("Content-Type")

	err = minIO.UploadFileFromReader(h.Client, bucketName, objectKey, file, fileHeader.Size, contentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	resizeTo := c.PostForm("resize_to")
	watermarkText := c.PostForm("watermark_text")
	getThumbnail := c.PostForm("get_thumbnail") == "true"

	photo := &repository.Photo{
		UUID:          &imageUUID,
		Status:        "в обработке",
		ResizeTo:      resizeTo,
		WatermarkText: watermarkText,
		GenThumbnail:  getThumbnail,
	}

	err = repository.InsertPhotoData(h.DB, photo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	err = startKafka.SendMessage(photo, h.Writer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ginext.H{"Photo put in queue": imageUUID})
}

func (h *Handler) GetPhoto(c *ginext.Context) {
	id := c.Param("id")

	bucketName := "images"
	file, err := minIO.GetPhoto(h.Client, bucketName, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	objInfo, err := file.Stat()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	imageData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, objInfo.ContentType, imageData)
}

func (h *Handler) DeletePhoto(c *ginext.Context) {
	id := c.Param("id")

	bucketName := "images"
	err := minIO.RemovePhoto(h.Client, bucketName, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ginext.H{"Photo deleted": id})
}

func (h *Handler) GetPhotoStatus(c *ginext.Context) {
	id := c.Param("id")

	status, err := repository.GetStatus(h.DB, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ginext.H{"status": status})
}
