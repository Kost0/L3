package repository

import (
	"github.com/google/uuid"
)

type Photo struct {
	UUID          *uuid.UUID `json:"uuid"`
	Status        string     `json:"status"`
	ResizeTo      string     `json:"resize_to"`
	WatermarkText string     `json:"watermark_text"`
	GenThumbnail  bool       `json:"gen_thumbnail"`
}
