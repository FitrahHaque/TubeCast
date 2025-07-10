package rss

import (
	"context"
	"fmt"
	"os"
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

func (user *User) createMetaStation(title string, description string) (MetaStation, error) {
	if StationNames.Has(title) {
		return MetaStation{}, fmt.Errorf("there already exists a station named `%s`. Try again with a different name.\n", title)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	coverImage, _ := Megh.upload(ctx, "", title, COVER)
	metaStation := MetaStation{
		ID:             uuid.New(),
		Title:          title,
		Description:    description,
		ChannelCount:   0,
		CreatedOn:      time.Now(),
		Language:       "English",
		Copyright:      user.Username,
		ITunesAuthor:   user.Name,
		ITunesSubtitle: "",
		ITunesSummary:  description,
		ITunesImage: ITunesImage{
			Href: coverImage,
		},
		ITunesExplicit: "no",
		ITunesCategories: []Category{
			{
				Text: "Technology",
			},
		},
		Owner: ITunesOwner{
			Name:  user.Name,
			Email: user.EmailId,
		},
	}
	StationNames.Add(title)
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

func getMetaStation(title string) (MetaStation, error) {
	if !StationNames.Has(title) {
		return MetaStation{}, fmt.Errorf("station `%s` does not exist\n", title)
	}
	return loadMetaStationFromLocal(Megh.getLocalStationFilepath(title))
}

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

func (station *Station) makeSpace(ctx context.Context, size uint64) bool {
	metaStation, err := getMetaStation(station.Title)
	if err != nil {
		return false
	}
	for {
		usage, err := Megh.getUsage(ctx)
		if err != nil {
			return false
		}
		// fmt.Println(used+sizeInBytes, MaximumStorage)
		if usage.TotalSizeBytes+size < Megh.MaximumStorage {
			// return "", fmt.Errorf("quota exceeded\n")
			return true
		} else if !station.removeOldestItem(&metaStation) {
			return false
		}
	}
}

func (station *Station) removeOldestItem(metaStation *MetaStation) bool {
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
	audioPath := Megh.getLocalAudioFilepath(id, station.Title)
	thumbnailPath := Megh.getLocalThumbnailFilepath(id, station.Title)
	if err := os.Remove(audioPath); err != nil {
		fmt.Printf("error-1: %v\n", err)
		return false
	}
	if err := os.Remove(thumbnailPath); err != nil {
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
