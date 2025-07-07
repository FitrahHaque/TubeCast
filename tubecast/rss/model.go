package rss

import (
	"time"

	"github.com/google/uuid"
)

type Station struct {
	ID               uuid.UUID     `xml:"id"                     json:"id"`
	Name             string        `xml:"name"                   json:"name"`
	Url              string        `xml:"url"                    json:"url"`
	Description      string        `xml:"description"            json:"description"`
	Items            []StationItem `xml:"items"                  json:"items"`
	Language         string        `xml:"language"               json:"language"`
	Copyright        string        `xml:"copyright"              json:"copyright"`
	ITunesAuthor     string        `xml:"itunes:author"          json:"itunes_author"`
	ITunesSubtitle   string        `xml:"itunes:subtitle"        json:"itunes_subtitle"`
	ITunesSummary    string        `xml:"itunes:summary"         json:"itunes_summary"`
	ITunesImage      string        `xml:"itunes:image,attr"      json:"itunes_image"`
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
	ID                uuid.UUID `xml:"id,attr"                      json:"id"`
	Title             string    `xml:"title"                        json:"title"`
	Enclosure         Enclosure `xml:"enclosure"                    json:"enclosure"`
	GUID              string    `xml:"guid"                         json:"guid"`
	Description       string    `xml:"description"                  json:"description"`
	PubDate           string    `xml:"pubDate"                      json:"pubDate"`
	ITunesDuration    string    `xml:"itunes:duration"              json:"itunes_duration"`
	ITunesExplicit    string    `xml:"itunes:explicit"              json:"itunes_explicit"`
	ITunesEpisode     int       `xml:"itunes:episode,omitempty"     json:"itunes_episode"`
	ITunesSeason      int       `xml:"itunes:season,omitempty"      json:"itunes_season"`
	ITunesEpisodeType string    `xml:"itunes:episodeType"           json:"itunes_episode_type"`
	ThumbnailUrl      string    `xml:"thumbnail_url"                json:"thumbnail_url"`
}

type Enclosure struct {
	URL    string `xml:"url,attr"    json:"enclosure_url"`
	Length uint64 `xml:"length,attr" json:"enclosure_length"`
	Type   string `xml:"type,attr"   json:"enclosure_type"`
}

type MetaStation struct {
	ID               uuid.UUID         `json:"id"`
	Name             string            `json:"name"`
	Url              string            `json:"url"`
	Description      string            `json:"description"`
	Items            []MetaStationItem `json:"items"`
	ChannelCount     uint32            `json:"channel_count"`
	CreatedOn        time.Time         `json:"created_on"`
	Language         string            `json:"language"`
	Copyright        string            `json:"copyright"`
	ITunesAuthor     string            `json:"itunes_author"`
	ITunesSubtitle   string            `json:"itunes_subtitle"`
	ITunesSummary    string            `json:"itunes_summary"`
	ITunesImage      string            `json:"itunes_image"`
	ITunesExplicit   string            `json:"itunes_explicit"`
	ITunesCategories []Category        `json:"itunes_categories"`
	Owner            ITunesOwner       `json:"itunes_owner"`
}

type MetaStationItem struct {
	ID                uuid.UUID `json:"id"`
	Title             string    `json:"title"`
	Description       string    `json:"description"`
	Author            string    `json:"author"`
	ChannelID         string    `json:"channel_id"`
	Views             uint32    `json:"views"`
	AddedOn           time.Time `json:"added_on"`
	Enclosure         Enclosure `json:"enclosure"`
	GUID              string    `json:"guid"`
	PubDate           string    `json:"pubDate"`
	ITunesDuration    string    `json:"itunes_duration"`
	ITunesExplicit    string    `json:"itunes_explicit"`
	ITunesEpisode     int       `json:"itunes_episode"`
	ITunesSeason      int       `json:"itunes_season"`
	ITunesEpisodeType string    `json:"itunes_episode_type"`
	ThumbnailUrl      string    `json:"thumbnail_url"`
}

type User struct {
	Name      string
	YouTubeID string
	AppleID   string
	Playlists []Playlist
}

type Playlist struct {
	Name         string
	Url          string
	LastUpdated  time.Time
	StationItems []StationItem
}
