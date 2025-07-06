package rss

import (
	"time"

	"github.com/google/uuid"
)

type Station struct {
	ID          uuid.UUID     `xml:"id"          json:"id"`
	Name        string        `xml:"name"        json:"name"`
	Url         string        `xml:"url"         json:"url"`
	Description string        `xml:"description" json:"description"`
	Items       []StationItem `xml:"items"       json:"items"`
}

type StationItem struct {
	ID          uuid.UUID `xml:"id"          json:"id"`
	Name        string    `xml:"name"        json:"name"`
	Link        string    `xml:"link"        json:"link"`
	Description string    `xml:"description" json:"description"`
	UploadedOn  time.Time `xml:"uploaded_on" json:"uploaded_on"`
}

type MetaStation struct {
	ID           uuid.UUID         `json:"id"`
	Name         string            `json:"name"`
	Url          string            `json:"url"`
	Description  string            `json:"description"`
	Items        []MetaStationItem `json:"items"`
	ChannelCount uint32            `json:"channel_count"`
}

type MetaStationItem struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Link        string    `json:"link"`
	Description string    `json:"description"`
	UploadedOn  time.Time `json:"uploaded_on"`
	ChannelID   string    `json:"channel_id"`
	Views       uint32    `json:"views"`
}
