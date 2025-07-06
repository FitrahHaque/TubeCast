package rss

import (
	"time"

	"github.com/google/uuid"
)

type Station struct {
	ID          uuid.UUID     `xml:"id"`
	Name        string        `xml:"name"`
	Url         string        `xml:"url"`
	Description string        `xml:"description"`
	Items       []StationItem `xml:"items"`
}

type StationItem struct {
	ID          uuid.UUID `xml:"id"`
	Name        string    `xml:"name"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	UploadedOn  time.Time `xml:"uploaded_on"`
}

type metaStation struct {
	station      *Station
	channelCount uint32
	items        []metaStationItem
}

type metaStationItem struct {
	stationItem *StationItem
	channelId   string
	views       uint32
}

type Set[T comparable] struct {
	set map[T]struct{}
}
