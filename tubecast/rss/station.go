package rss

import (
	"fmt"
	"path/filepath"
	"time"

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
			Url:              metaStation.Url,
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
	station.saveXMLToLocal()
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
			Url:              metaStation.Url,
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
		Title:          metaItem.Title,
		Enclosure:      metaItem.Enclosure,
		Description:    metaItem.Description,
		GUID:           metaItem.GUID,
		PubDate:        metaItem.PubDate,
		ITunesDuration: metaItem.ITunesDuration,
		ITunesExplicit: metaItem.ITunesExplicit,
		ITunesImage:    metaItem.ITunesImage,
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
