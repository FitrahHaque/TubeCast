package rss

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func SyncChannel(title, description string, channelID string) (string, error) {
	metaStation, err := getMetaStation(title, description)
	if err != nil {
		return "", err
	}
	metaStation.subscribeToChannel(channelID)
	return metaStation.syncChannel(channelID)
}

func CreateShow(title, description string) string {
	if !StationNames.Has(title) {
		Usr.createMetaStation(title, description)
	}
	return Megh.getShareableFeedUrl(title)
}

func Sync() error {
	for title := range StationNames.Value {
		metaStation, err := getMetaStation(title, "")
		if err != nil {
			return err
		}
		for channel := range metaStation.SubscribedChannel.Value {
			if _, err := metaStation.syncChannel(channel); err != nil {
				return err
			}
		}
	}
	return nil
}

func RemoveVideoFromShow(title, videoUrl string) error {
	if !StationNames.Has(title) {
		return errors.New("show with this title does not exist")
	}
	metaStation, err := getMetaStation(title, "")
	if err != nil {
		return err
	}
	return metaStation.deleteVideo(videoUrl)
}

func RemoveShow(title string) error {
	if !StationNames.Has(title) {
		return errors.New("show with this title does not exist")
	}
	metaStation, err := getMetaStation(title, "")
	if err != nil {
		return err
	}
	return metaStation.delete()
}

func AddVideoToShow(title, description string, videoUrl string) (string, error) {
	metaStation, err := getMetaStation(title, description)
	if err != nil {
		return "", err
	}
	return metaStation.addVideo(videoUrl)
}

func (station *Station) Print() {
	fmt.Println("------ Station Print ------")
	fmt.Printf("ID: %v\n", station.ID)
	fmt.Printf("Name: %v\n", station.Title)
	fmt.Printf("Description: %v\n", station.Description)
	fmt.Printf("Language: %v\n", station.Language)
	fmt.Printf("Copyright: %v\n", station.Copyright)
	fmt.Printf("ITunnesAuthor: %v\n", station.ITunesAuthor)
	fmt.Printf("ITunesSubtitle: %v\n", station.ITunesSubtitle)
	fmt.Printf("ITunesSummary: %v\n", station.ITunesSummary)
	fmt.Printf("ITunesImage: %v\n", station.ITunesImage)
	fmt.Printf("ITunesExplicit: %v\n", station.ITunesExplicit)
	fmt.Printf("ITunesCategories: %v\n", station.ITunesCategories)
	for i, item := range station.Items {
		fmt.Printf("Item %v:\n", i)
		item.Print()
	}
	fmt.Println("----------------- ---------------------")
}

func (stationItem *StationItem) Print() {
	fmt.Println("--------- Station Item ---------")
	fmt.Printf("Title: %v\n", stationItem.Title)
	fmt.Printf("Enclosure: %v\n", stationItem.Enclosure)
	fmt.Printf("GUID: %v\n", stationItem.GUID)
	fmt.Printf("Description: %v\n", stationItem.Description)
	fmt.Printf("PubDate: %v\n", stationItem.PubDate)
	fmt.Printf("ITunesDuration: %v\n", stationItem.ITunesDuration)
	fmt.Printf("ITunesExplicit: %v\n", stationItem.ITunesExplicit)
	// fmt.Printf("ITunesEpisode: %v\n", stationItem.ITunesEpisode)
	// fmt.Printf("ITunesSeason: %v\n", stationItem.ITunesSeason)
	// fmt.Printf("ITunesEpisodeType: %v\n", stationItem.ITunesEpisodeType)
	fmt.Println("----------------- ---------------------")
}

func Init() {
	Usr = User{
		Username: os.Getenv("USERNAME"),
	}
	Usr.Username = strings.ToLower(Usr.Username)
	// fmt.Printf("username: %v\n", Usr.Username)
	Megh = Cloud{
		ArchiveId:        Usr.getArchiveIdentifier(),
		FeedUrlPrefix:    Usr.getFeedUrlPrefix(),
		MaximumStorage:   10 * 1024 * 1024 * 1024, //10 GiB
		ArchiveUrlPrefix: fmt.Sprintf("https://archive.org/download/%s/", Usr.getArchiveIdentifier()),
	}
	isArch := os.Getenv("ARCHIVE")
	if isArch == "Yes" {
		Megh.IsArchive = true
	}
	if err := loadAllMetaStationNames(); err != nil {
		fmt.Printf("error in init: %v\n", err)
	}
}
