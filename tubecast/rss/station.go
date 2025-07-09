package rss

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/google/uuid"
)

func (user *User) CreateStation(name, description string) (Station, error) {
	if metaStation, err := user.createMetaStation(name, description); err != nil {
		return Station{}, err
	} else {
		station := Station{
			ID:               metaStation.ID,
			Title:            metaStation.Title,
			Description:      metaStation.Description,
			Items:            getStationItems(metaStation.Items),
			Language:         metaStation.Language,
			Copyright:        metaStation.Copyright,
			ITunesAuthor:     metaStation.ITunesAuthor,
			ITunesSubtitle:   metaStation.ITunesSubtitle,
			ITunesSummary:    metaStation.ITunesSummary,
			ITunesImage:      metaStation.ITunesImage,
			ITunesExplicit:   metaStation.ITunesExplicit,
			ITunesCategories: metaStation.ITunesCategories,
			Owner:            metaStation.Owner,
		}
		return station, nil
	}
}

// func (station *Station) AddToStation(
// 	title string,
// 	description string,
// 	link string,
// 	uploadedOn string,
// 	views uint32,
// 	author string,
// 	length uint64,
// ) error {
// 	metaStation, err := getMetaStation(station.Name)
// 	if err != nil {
// 		return err
// 	}

// 	metaItem := metaStation.addToStation(
// 		title,
// 		description,
// 		link,
// 		author,
// 		uploadedOn,
// 		views,
// 		length,
// 	)
// 	item := getStationItem(metaItem)
// 	station.Items = append(station.Items, item)
// 	return nil
// }

func (station *Station) addToStation(stationItem StationItem) {
	station.Items = append(station.Items, stationItem)
}

func GetStation(name string) (Station, error) {
	if !StationNames.Has(name) {
		var user User //change this
		return user.CreateStation(name, "name:name")
	}
	if metaStation, err := getMetaStation(name); err != nil {
		return Station{}, err
	} else {
		station := Station{
			ID:               metaStation.ID,
			Title:            metaStation.Title,
			Description:      metaStation.Description,
			Items:            getStationItems(metaStation.Items),
			Language:         metaStation.Language,
			Copyright:        metaStation.Copyright,
			ITunesAuthor:     metaStation.ITunesAuthor,
			ITunesSubtitle:   metaStation.ITunesSummary,
			ITunesSummary:    metaStation.ITunesSummary,
			ITunesImage:      metaStation.ITunesImage,
			ITunesExplicit:   metaStation.ITunesExplicit,
			ITunesCategories: metaStation.ITunesCategories,
			Owner:            metaStation.Owner,
		}
		return station, nil
	}
}

func (user *User) createMetaStation(name string, description string) (MetaStation, error) {
	if StationNames.Has(name) {
		return MetaStation{}, fmt.Errorf("there already exists a station named `%s`. Try again with a different name.\n", name)
	}
	coverImage, err := getCoverImage(name)
	if err != nil {
		return MetaStation{}, err
	}
	metaStation := MetaStation{
		ID:             uuid.New(),
		Title:          name,
		Description:    description,
		ChannelCount:   0,
		CreatedOn:      time.Now(),
		Language:       "English",
		Copyright:      user.YouTubeID,
		ITunesAuthor:   user.Name,
		ITunesSubtitle: "",
		ITunesSummary:  description,
		ITunesImage:    coverImage,
		ITunesExplicit: "no",
		ITunesCategories: []Category{
			{
				Text: "Technology",
			},
		},
		Owner: ITunesOwner{
			Name:  user.Name,
			Email: user.AppleID,
		},
	}
	StationNames.Add(name)
	return metaStation, metaStation.saveMetaStationToLocal()
}

func getStationItems(metaItems []MetaStationItem) []StationItem {
	items := make([]StationItem, len(metaItems))
	for i, metaItem := range metaItems {
		items[i] = getStationItem(metaItem)
	}
	return items
}

func getStationItem(metaItem MetaStationItem) StationItem {
	return StationItem{
		GUID:           metaItem.GUID,
		Title:          metaItem.Title,
		Enclosure:      metaItem.Enclosure,
		ITunesImage:    metaItem.ITunesImage,
		Description:    metaItem.Description,
		Link:           metaItem.Link,
		PubDate:        metaItem.PubDate,
		ITunesDuration: metaItem.ITunesDuration,
		ITunesExplicit: metaItem.ITunesExplicit,
		ITunesAuthor:   metaItem.ITunesAuthor,
		ITunesSubtitle: metaItem.ITunesSubtitle,
		ITunesSummary:  metaItem.ITunesSummary,
		// ITunesEpisode:     metaItem.ITunesEpisode,
		// ITunesSeason:      metaItem.ITunesSeason,
		// ITunesEpisodeType: metaItem.ITunesEpisodeType,
	}
}

func getMetaStation(name string) (MetaStation, error) {
	if !StationNames.Has(name) {
		return MetaStation{}, fmt.Errorf("station `%s` does not exist\n", name)
	}
	return loadMetaStationFromLocal(filepath.Join(STATION_BASE, name+".json"))
}

// func (station *MetaStation) addToStation(
//
//	title,
//	description,
//	link,
//	author string,
//	uploadedOn string,
//	views uint32,
//	length uint64,
//
//	) MetaStationItem {
//		id := uuid.New()
//		item := MetaStationItem{
//			ID:          id,
//			Title:       title,
//			Description: description,
//			Author:      author,
//			Views:       views,
//			AddedOn:     time.Now(),
//			Enclosure: Enclosure{
//				URL:    link,
//				Length: length,
//				Type:   "audio/mpeg",
//			},
//			GUID:           "",
//			PubDate:        uploadedOn,
//			ITunesDuration: fmt.Sprint(length),
//			ITunesExplicit: "no",
//		}
//		station.Items = append(station.Items, item)
//		return item
//	}

func (metaStation *MetaStation) addToStation(stationItem MetaStationItem) {
	metaStation.Items = append(metaStation.Items, stationItem)
	metaStation.saveMetaStationToLocal()
}

func (station *Station) HasItem(id string) bool {
	for _, item := range station.Items {
		if item.GUID == id {
			return true
		}
	}
	return false
}

func (station *Station) GetStationItem(id string) (StationItem, bool) {
	for _, item := range station.Items {
		if item.GUID == id {
			return item, true
		}
	}
	return StationItem{}, false
}

func (station *Station) filter(ids []string) []string {
	var videoIds []string
	for _, id := range ids {
		if !station.HasItem(id) {
			videoIds = append(videoIds, id)
		}
	}
	return videoIds
}

func (station *Station) makeSpace(sizeInBytes uint64) bool {
	accessToken, err := TOKEN_MANAGER.GetValidAccessToken()
	if err != nil {
		return false
	}
	cfg := dropbox.Config{Token: accessToken}
	dbx := files.New(cfg)
	metaStation, err := getMetaStation(station.Title)
	if err != nil {
		return false
	}
	for {
		used, err := dirSize(dbx, DROPBOX_AUDIO_BASE)
		if err != nil {
			return false
		}
		fmt.Println(used+sizeInBytes, MaximumCloudStorage)
		if used+sizeInBytes < MaximumCloudStorage {
			// return "", fmt.Errorf("quota exceeded\n")
			return true
		} else if !station.removeOldestItem(accessToken, &metaStation) {
			return false
		}
	}
}

func (station *Station) removeOldestItem(accessToken string, metaStation *MetaStation) bool {
	if len(station.Items) == 0 {
		return false
	}
	oldest := metaStation.Items[0].AddedOn
	oldestIndex := 0
	for i, item := range metaStation.Items {
		if item.AddedOn.Before(oldest) {
			oldest = item.AddedOn
			oldestIndex = i
		}
	}
	id := metaStation.Items[oldestIndex].GUID
	audioPath := filepath.Join(DROPBOX_AUDIO_BASE, station.Title, id+".mp3")
	thumbnailPath := filepath.Join(DROPBOX_THUMBNAILS_BASE, station.Title, id+".png")
	if err := deleteDropboxFile(accessToken, audioPath); err != nil {
		fmt.Printf("error-1: %v\n", err)
		return false
	}
	if err := deleteDropboxFile(accessToken, thumbnailPath); err != nil {
		fmt.Printf("error-2: %v\n", err)
		return false
	}
	if station.Items[oldestIndex].GUID == metaStation.Items[oldestIndex].GUID {
		metaStation.Items = append(metaStation.Items[:oldestIndex], metaStation.Items[oldestIndex+1:]...)
		station.Items = append(station.Items[:oldestIndex], station.Items[oldestIndex+1:]...)
		station.updateFeed()
	} else {
		fmt.Println("not in order with meta")
	}
	fmt.Printf("file deleted with id %v\n", id)
	return true
}
