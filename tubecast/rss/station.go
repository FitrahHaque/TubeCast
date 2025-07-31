package rss

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

func getMetaStation(title string, description string) (MetaStation, error) {
	if !StationNames.Has(title) {
		return Usr.createMetaStation(title, description)
	}
	return loadMetaStationFromLocal(Megh.getLocalStationFilepath(title))
}

func (user *User) createMetaStation(title string, description string) (MetaStation, error) {
	if StationNames.Has(title) {
		return MetaStation{}, fmt.Errorf("there already exists a station named `%s`. Try again with a different name.\n", title)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	localpath2 := Megh.getLocalCoverFilepath(title)
	localpath1 := strings.Split(localpath2, ".")[0] + ".webp"
	ConvertImageToCorrectFormat(localpath1, localpath2)
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
		SubscribedChannel: NewSet[string](),
	}
	StationNames.Add(title)
	metaStation.updateFeed()
	return metaStation, nil
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

// func getMetaStation(title string) (MetaStation, error) {
// 	if !StationNames.Has(title) {
// 		return MetaStation{}, fmt.Errorf("station `%s` does not exist\n", title)
// 	}
// 	return
// }

func (metaStation *MetaStation) addToStation(stationItem MetaStationItem) {
	metaStation.Items = append(metaStation.Items, stationItem)
	// metaStation.
	metaStation.updateFeed()
}

func (metaStation *MetaStation) subscribeToChannel(channelId string) {
	metaStation.SubscribedChannel.Add(channelId)
	metaStation.updateFeed()
}

func (metaStation *MetaStation) HasItem(id string) bool {
	for _, item := range metaStation.Items {
		if item.GUID == id {
			return true
		}
	}
	return false
}

func (metaStation *MetaStation) getAllItems() []EpisodeInfo {
	var out []EpisodeInfo
	for _, item := range metaStation.Items {
		out = append(out, EpisodeInfo{
			Title:  item.Title,
			Author: item.ITunesAuthor,
		})
	}
	return out
}

// func (station *Station) GetStationItem(id string) (StationItem, bool) {
// 	for _, item := range station.Items {
// 		if item.GUID == id {
// 			return item, true
// 		}
// 	}
// 	return StationItem{}, false
// }

func (metaStation *MetaStation) filter(ids []string) []string {
	var videoIds []string
	for _, id := range ids {
		if !metaStation.HasItem(id) {
			videoIds = append(videoIds, id)
		}
	}
	return videoIds
}

func (metaStation *MetaStation) makeSpace(ctx context.Context, size uint64) bool {
	for {
		usage, err := Megh.getUsage(ctx)
		if err != nil {
			return false
		}
		// fmt.Println(used+sizeInBytes, MaximumStorage)
		if usage.TotalSizeBytes+size < Megh.MaximumStorage {
			// return "", fmt.Errorf("quota exceeded\n")
			return true
		} else if !metaStation.removeOldestItem(ctx) {
			return false
		}
	}
}

func (metaStation *MetaStation) removeOldestItem(ctx context.Context) bool {
	if len(metaStation.Items) == 0 {
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
	if err := Megh.deleteEpisode(ctx, id, metaStation.Title); err != nil {
		fmt.Printf("error-1: %v\n", err)
		return false
	}
	metaStation.Items = append(metaStation.Items[:oldestIndex], metaStation.Items[oldestIndex+1:]...)
	metaStation.updateFeed()
	fmt.Printf("file deleted with id %v\n", id)
	return true
}
