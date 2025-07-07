package rss

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func (station *Station) SyncChannel(ChannelFeedUrl string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	ids, err := getLatestVideos(ctx, ChannelFeedUrl, 2)
	if err != nil {
		return err
	}

	fmt.Println("Latest video urls: ", ids)
	metaStation, err := getMetaStation(station.Name)
	if err != nil {
		return err
	}
	for _, id := range ids {
		var wg sync.WaitGroup
		metaStationItem := MetaStationItem{
			ID:        id,
			AddedOn:   time.Now(),
			ChannelID: ChannelFeedUrl,
			Author:    "ThePrimeagen",
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			if title, err := getVideoTitle(ctx, id); err != nil {
				return
			} else {
				metaStationItem.Title = title
			}
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			if description, err := getVideoDescription(ctx, id); err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			} else {
				metaStationItem.Description = description
			}
		}()
		wg.Add(1)
		go func() {
			wg.Done()
			if duration, err := getVideoDuration(ctx, id); err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			} else {
				metaStationItem.ITunesDuration = duration
			}
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			if views, err := getVideoViews(ctx, id); err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			} else {
				metaStationItem.Views = views
			}
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			if pubDate, err := getVideoPubDate(ctx, id); err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			} else {
				metaStationItem.PubDate = pubDate
			}
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := metaStationItem.saveVideoThumbnail(ctx, id); err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			} else {
				localPath := filepath.Join(THUMBNAILS_BASE, metaStationItem.ID+".webp")
				// fmt.Printf("localPath: %v\n", localPath)
				if share, err := uploadToDropbox(localPath, filepath.Join(DROPBOX_THUMBNAILS_BASE, metaStationItem.ID+".webp")); err != nil {
					fmt.Printf("Error: %v\n", err)
					return
				} else {
					fmt.Println(share)
					metaStationItem.ThumbnailUrl = share
					os.Remove(localPath)
				}
			}
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := metaStationItem.saveAudio(ctx, id); err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			} else {
				localPath := filepath.Join(AUDIO_BASE, metaStationItem.ID+".mp3")
				if share, err := uploadToDropbox(localPath, filepath.Join(DROPBOX_AUDIO_BASE, metaStationItem.ID+".mp3")); err != nil {
					fmt.Printf("Error: %v\n", err)
					return
				} else {
					metaStationItem.Enclosure = Enclosure{
						URL:  share,
						Type: "audio/mpeg",
					}
					os.Remove(localPath)
				}
			}
		}()
		wg.Wait()
		metaStation.addToStation(metaStationItem)
		stationItem := getStationItem(metaStationItem)
		station.addToStation(stationItem)
	}
	return nil
}

func (station *Station) Print() {
	fmt.Println("------ Station Print ------")
	fmt.Printf("ID: %v\n", station.ID)
	fmt.Printf("Name: %v\n", station.Name)
	fmt.Printf("Url: %v\n", station.Url)
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
	fmt.Printf("ID: %v\n", stationItem.ID)
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
	fmt.Printf("ThumbnailUrl: %v\n", stationItem.ThumbnailUrl)
	fmt.Println("----------------- ---------------------")
}

func Init() {
	TOKEN_MANAGER = NewTokenManager(
		os.Getenv("DROPBOX_APP_KEY"),
		os.Getenv("DROPBOX_APP_SECRET"),
		os.Getenv("DROPBOX_REFRESH_TOKEN"),
	)
	loadAllMetaStationNames()
	for i, _ := range StationNames.set {
		fmt.Printf("key: %v\n", i)
	}
}
