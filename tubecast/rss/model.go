package rss

import (
	"encoding/xml"
	"time"

	"github.com/google/uuid"
)

type Station struct {
	XMLName          xml.Name      `xml:"channel"                json:"channel"`
	ID               uuid.UUID     `xml:"id"                     json:"id"`
	Title            string        `xml:"title"                  json:"title"`
	ITunesImage      ITunesImage   `xml:"itunes:image"           json:"itunes_image"`
	Description      string        `xml:"description"            json:"description"`
	Items            []StationItem `xml:"item"                   json:"item"`
	Language         string        `xml:"language"               json:"language"`
	Copyright        string        `xml:"copyright"              json:"copyright"`
	ITunesAuthor     string        `xml:"itunes:author"          json:"itunes_author"`
	ITunesSubtitle   string        `xml:"itunes:subtitle"        json:"itunes_subtitle"`
	ITunesSummary    string        `xml:"itunes:summary"         json:"itunes_summary"`
	ITunesExplicit   string        `xml:"itunes:explicit"        json:"itunes_explicit"`
	ITunesCategories []Category    `xml:"itunes:category"        json:"itunes_categories"`
	Owner            ITunesOwner   `xml:"itunes:owner"           json:"itunes_owner"`
}

type Category struct {
	Text        string    `xml:"text,attr"                 json:"text"`
	Subcategory *Category `xml:"itunes:category,omitempty" json:"subcategory,omitempty"`
}

type ITunesOwner struct {
	Name  string `xml:"itunes:name"  json:"name"`
	Email string `xml:"itunes:email" json:"email"`
}

type StationItem struct {
	// ID             string      `xml:"id,attr"                      json:"id"`
	// ITunesTitle    string      `xml:"itunes:title"                 json:"itunes_title"`
	GUID           string      `xml:"guid"                         json:"guid"`
	Title          string      `xml:"title"                        json:"title"`
	Enclosure      Enclosure   `xml:"enclosure"                    json:"enclosure"`
	ITunesImage    ITunesImage `xml:"itunes:image"                 json:"itunes_image"`
	Description    string      `xml:"description"                  json:"description"`
	Link           string      `xml:"link"                         json:"link"`
	PubDate        string      `xml:"pubDate"                      json:"pubDate"`
	ITunesDuration string      `xml:"itunes:duration"              json:"itunes_duration"`
	ITunesExplicit string      `xml:"itunes:explicit"              json:"itunes_explicit"`
	ITunesAuthor   string      `xml:"itunes:author"                json:"itunes_author"`
	ITunesSubtitle string      `xml:"itunes:subtitle"              json:"itunes_subtitle"`
	ITunesSummary  string      `xml:"itunes:summary"               json:"itunes_summary"`
	// ITunesEpisode     int       `xml:"itunes:episode,omitempty"     json:"itunes_episode"`
	// ITunesSeason      int       `xml:"itunes:season,omitempty"      json:"itunes_season"`
	// ITunesEpisodeType string    `xml:"itunes:episodeType"           json:"itunes_episode_type"`
}

type ITunesImage struct {
	Href string `xml:"href,attr" json:"itunes_image_href"`
}
type Enclosure struct {
	URL    string `xml:"url,attr"    json:"enclosure_url"`
	Length uint64 `xml:"length,attr" json:"enclosure_length"`
	Type   string `xml:"type,attr"   json:"enclosure_type"`
}

type MetaStation struct {
	ID                uuid.UUID         `json:"id"`
	Title             string            `json:"title"`
	Url               string            `json:"url"`
	Description       string            `json:"description"`
	Items             []MetaStationItem `json:"item"`
	ChannelCount      uint32            `json:"channel_count"`
	CreatedOn         time.Time         `json:"created_on"`
	Language          string            `json:"language"`
	Copyright         string            `json:"copyright"`
	ITunesAuthor      string            `json:"itunes_author"`
	ITunesSubtitle    string            `json:"itunes_subtitle"`
	ITunesSummary     string            `json:"itunes_summary"`
	ITunesImage       ITunesImage       `json:"itunes_image"`
	ITunesExplicit    string            `json:"itunes_explicit"`
	ITunesCategories  []Category        `json:"itunes_categories"`
	Owner             ITunesOwner       `json:"itunes_owner"`
	SubscribedChannel *Set[string]      `json:"subscribed_channel"`
}

type MetaStationItem struct {
	GUID           string      `json:"guid"`
	ITunesAuthor   string      `json:"itunes_author"`
	ChannelID      string      `json:"channel_id"`
	AddedOn        time.Time   `json:"added_on"`
	ITunesExplicit string      `json:"itunes_explicit"`
	Title          string      `json:"title"`
	Description    string      `json:"description"`
	ITunesSummary  string      `json:"itunes_summary"`
	ITunesDuration string      `json:"itunes_duration"`
	Views          uint32      `json:"views"`
	PubDate        string      `json:"pubDate"`
	ITunesImage    ITunesImage `json:"itunes_image"`
	Enclosure      Enclosure   `json:"enclosure"`
	ITunesSubtitle string      `json:"itunes_subtitle"`
	Link           string      `json:"link"`
	// ITunesEpisode     int       `json:"itunes_episode"`
	// ITunesSeason      int       `json:"itunes_season"`
	// ITunesEpisodeType string    `json:"itunes_episode_type"`
}

type User struct {
	EmailId   string
	Username  string // will be used as prefix in identifier
	Name      string
	Playlists []Playlist
}

type Playlist struct {
	Name         string
	Url          string
	LastUpdated  time.Time
	StationItems []StationItem
}

type Cloud struct {
	ArchiveId        string
	FeedUrlPrefix    string
	MaximumStorage   uint64
	ArchiveUrlPrefix string
	// IAS3AccessKey    string
	// IAS3SecretKey    string
}

type Usage struct {
	TotalSizeBytes uint64 `json:"total_size_bytes"`
	TotalSizeMiB   uint64 `json:"total_size_mib"`
	FileCount      uint64 `json:"file_count"`
}
