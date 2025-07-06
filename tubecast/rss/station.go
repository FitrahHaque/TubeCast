package rss

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

func CreateStation(name, description string) (Station, error) {
	if metaStation, err := createMetaStation(name, description); err != nil {
		return Station{}, err
	} else {
		station := Station{
			ID:          metaStation.ID,
			Name:        metaStation.Name,
			Description: metaStation.Description,
			Url:         metaStation.Url,
			Items:       getStationItems(metaStation.Items),
		}
		return station, nil
	}
}

func (station *Station) AddToStation(
	name string,
	description string,
	link string,
	uploadedOn time.Time,
	views uint32,
	channelID string,
) error {
	metaStation, err := getMetaStation(station.Name)
	if err != nil {
		return err
	}

	id := metaStation.addToStation(
		name,
		description,
		link,
		channelID,
		uploadedOn,
		views,
	)
	item := StationItem{
		ID:          id,
		Name:        name,
		Description: description,
		Link:        link,
		UploadedOn:  uploadedOn,
	}
	station.Items = append(station.Items, item)
	return nil
}

func GetStation(name string) (Station, error) {
	if !StationNames.Has(name) {
		return Station{}, fmt.Errorf("station `%s` does not exist\n", name)
	}
	if metaStation, err := getMetaStation(name); err != nil {
		return Station{}, err
	} else {
		station := Station{
			ID:          metaStation.ID,
			Name:        metaStation.Name,
			Url:         metaStation.Url,
			Description: metaStation.Description,
			Items:       getStationItems(metaStation.Items),
		}
		return station, nil
	}
}

func createMetaStation(name string, description string) (MetaStation, error) {
	if StationNames.Has(name) {
		return MetaStation{}, fmt.Errorf("there already exists a station named `%s`. Try again with a different name.\n", name)
	}

	metaStation := MetaStation{
		ID:           uuid.New(),
		Name:         name,
		Description:  description,
		ChannelCount: 0,
	}
	StationNames.Add(name)
	return metaStation, saveMetaStationToLocal(fmt.Sprintf("%s/%s.json", STATION_BASE, name), metaStation)
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
		ID:          metaItem.ID,
		Name:        metaItem.Name,
		Link:        metaItem.Link,
		Description: metaItem.Description,
		UploadedOn:  metaItem.UploadedOn,
	}
}

func getMetaStation(name string) (MetaStation, error) {
	if !StationNames.Has(name) {
		return MetaStation{}, fmt.Errorf("station `%s` does not exist\n", name)
	}
	return loadMetaStationFromLocal(fmt.Sprintf("%s/%s.json", STATION_BASE, name))
}

func (station *MetaStation) addToStation(
	name,
	description,
	link,
	channelID string,
	uploadedOn time.Time,
	views uint32,
) uuid.UUID {
	id := uuid.New()
	item := MetaStationItem{
		ID:          id,
		Name:        name,
		Description: description,
		Link:        link,
		ChannelID:   channelID,
		UploadedOn:  uploadedOn,
		Views:       views,
	}
	station.Items = append(station.Items, item)
	return id
}
