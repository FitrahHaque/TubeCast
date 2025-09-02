package rss

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func SyncChannel(title, channelID string) (string, error) {
	metaStation, err := getMetaStation(title, "")
	if err != nil {
		return "", err
	}
	metaStation.subscribeToChannel(channelID)
	return metaStation.syncChannel(channelID)
}

func CreateShow(title, description, coverFile string) (string, error) {
	if !StationNames.Has(title) {
		srcPath := filepath.Join(COVER_BASE, coverFile)
		destPath := Megh.getLocalCoverFilepath(title)
		destPath = strings.TrimSuffix(destPath, filepath.Ext(destPath)) + filepath.Ext(srcPath)
		if srcPath != destPath {
			srcFile, err := os.Open(srcPath)
			if err != nil {
				return "", err
			}
			if err = os.MkdirAll(COVER_BASE, 0o755); err != nil {
				return "", err
			}
			destFile, err := os.Create(destPath)
			if err != nil {
				return "", err
			}
			if _, err = io.Copy(destFile, srcFile); err != nil {
				return "", err
			}
		}
		if _, err := Usr.createMetaStation(title, description); err != nil {
			return "", err
		}
	}
	return Megh.getShareableFeedUrl(title), nil
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

func RemoveVideoFromShow(showTitle, videoTitle, author string) error {
	if !StationNames.Has(showTitle) {
		return errors.New("show with this title does not exist")
	}
	metaStation, err := getMetaStation(showTitle, "")
	if err != nil {
		return err
	}
	return metaStation.deleteVideo(videoTitle, author)
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

func AddVideoToShow(title, videoUrl string) (string, error) {
	metaStation, err := getMetaStation(title, "")
	if err != nil {
		return "", err
	}
	return metaStation.addVideo(videoUrl)
}

func GetAllShowEpisodes(title string) ([]EpisodeInfo, error) {
	if !StationNames.Has(title) {
		return nil, errors.New("Show with the title does not exist")
	}
	metaStation, err := getMetaStation(title, "")
	if err != nil {
		return nil, err
	}
	items := metaStation.getAllItems()
	sort.Slice(items, func(i, j int) bool {
		timeA, err := time.Parse(time.RFC1123, items[i].PubDate)
		if err != nil {
			logError(err, "Invalid Date conversion")
			return false
		}
		timeB, err := time.Parse(time.RFC1123, items[j].PubDate)
		if err != nil {
			logError(err, "Invalid Date Conversion")
		}
		return timeA.After(timeB)
	})
	return items, nil
}

func GetFeedUrl(title string) string {
	return Megh.getShareableFeedUrl(title)
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
		// fmt.Printf("error in init: %v\n", err)
	}
}
