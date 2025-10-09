package repository

import (
	"time"

	"github.com/google/uuid"
)

type URL struct {
	UUID     *uuid.UUID `json:"uuid"`
	URL      string     `json:"url"`
	ShortURL string     `json:"short_url"`
}

type URLInfo struct {
	UUID      *uuid.UUID `json:"uuid"`
	LinkID    *uuid.UUID `json:"link_id"`
	Time      time.Time  `json:"time"`
	UserAgent string     `json:"user_agent"`
	IP        string     `json:"ip"`
}

type AnalyticsGroups struct {
	Parameter      string `json:"parameter"`
	Visitors       int    `json:"visitors"`
	UniqueVisitors int    `json:"unique_visitors"`
}
