package rss

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

func (station *Station) SyncChannel(username string) (string, error) {
	metaStation, err := getMetaStation(station.Title)
	if err != nil {
		return "", errors.New("no metastation with this title")
	}
	channelFeedUrl, err := GetChannelFeedUrl(username)
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	ids, err := station.getLatestVideos(ctx, channelFeedUrl, 1)
	if err != nil {
		return "", err
	}
	fmt.Println("Latest video urls to be uploaded: ", ids)
	for _, id := range ids {
		if _, err := station.addItemToStation(ctx, id, username, channelFeedUrl); err != nil {
			log.Printf("error downloading video %s, error: %v\n", id, err)
		}
	}
	return metaStation.updateFeed()
}

func (station *Station) AddVideo(videoUrl string) (string, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	id, err := getVideoId(ctx, videoUrl)
	if err != nil {
		return "", err
	}
	ids := station.filter([]string{id})
	// fmt.Printf("ids...\n")
	if len(ids) == 0 {
		return "", errors.New("video already exists in the channel")
	}
	// fmt.Printf("username starting...\n")
	username, err := getVideoUsername(ctx, videoUrl)
	if err != nil {
		return "", err
	}
	channelFeedUrl, err := GetChannelFeedUrl(username)
	if err != nil {
		return "", err
	}
	if share, err := station.addItemToStation(ctx, id, username, channelFeedUrl); err != nil {
		return "", err
	} else {
		return share, nil
	}
}

func (station *Station) addItemToStation(ctx context.Context, id, username, channelFeedUrl string) (string, error) {
	metaStation, err := getMetaStation(station.Title)
	if err != nil {
		return "", errors.New("no metastation with this title")
	}
	metaStationItem := MetaStationItem{
		GUID:           id,
		ITunesAuthor:   username,
		ChannelID:      channelFeedUrl,
		AddedOn:        time.Now(),
		ITunesExplicit: "no",
		Link:           "https://www.youtube.com/watch?v=" + id,
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if title, err := getVideoTitle(ctx, metaStationItem.Link); err != nil {
			return
		} else {
			metaStationItem.Title = title
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		if description, err := getVideoDescription(ctx, metaStationItem.Link); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		} else {
			metaStationItem.Description = description
			metaStationItem.ITunesSubtitle = description
		}
	}()
	wg.Add(1)
	go func() {
		wg.Done()
		if duration, err := getVideoDuration(ctx, metaStationItem.Link); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		} else {
			metaStationItem.ITunesDuration = duration
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		if views, err := getVideoViews(ctx, metaStationItem.Link); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		} else {
			metaStationItem.Views = views
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		if pubDate, err := getVideoPubDate(ctx, metaStationItem.Link); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		} else {
			metaStationItem.PubDate = pubDate
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := metaStationItem.saveVideoThumbnail(ctx, station.Title, metaStationItem.Link); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		if size, err := metaStationItem.saveAudio(ctx, station.Title, metaStationItem.Link, 0); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		} else {
			station.makeSpace(ctx, size)
			if share, err := Megh.upload(ctx, metaStationItem.GUID, station.Title, AUDIO); err != nil {
				return
			} else {
				metaStationItem.Enclosure = Enclosure{
					URL:    share,
					Type:   "audio/mpeg",
					Length: size,
				}
				if share, err := Megh.upload(ctx, metaStationItem.GUID, station.Title, THUMBNAIL); err != nil {
					return
				} else {
					metaStationItem.ITunesImage = ITunesImage{
						Href: share,
					}
				}

			}
		}
	}()
	wg.Wait()
	if len(metaStationItem.Enclosure.URL) == 0 {
		return "", errors.New("could not upload audio")
	}
	metaStation.addToStation(metaStationItem)
	// stationItem := getStationItem(metaStationItem)
	// station.addToStation(stationItem)
	return metaStation.updateFeed()
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

func Init() User {
	var user User
	user.Username = os.Getenv("USERNAME")

	Megh = Cloud{
		ArchiveId:      user.getArchiveIdentifier(),
		FeedUrlPrefix:  user.getFeedUrlPrefix(),
		MaximumStorage: 2 * 1024 * 1024 * 1024, //2 GiB
		// MaximumStorage: 60 * 1024 * 1024,
	}
	Megh.ArchiveUrlPrefix = fmt.Sprintf("https://archive.org/download/%s/", Megh.ArchiveId)
	loadAllMetaStationNames()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	usage, err := Megh.getUsage(ctx)
	if err == nil {
		fmt.Printf("usage: %v\n", usage.TotalSizeMiB)
	}
	return user
}
