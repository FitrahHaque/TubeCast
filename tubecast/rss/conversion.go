package rss

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func (metaStation *MetaStation) syncChannel(channelUsername string) (string, error) {
	channelFeedUrl, err := GetChannelFeedUrl(channelUsername)
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	ids, err := metaStation.getLatestVideos(ctx, channelFeedUrl, 3)
	if err != nil {
		return "", err
	}
	// fmt.Println("Latest video urls to be uploaded: ", ids)
	for _, id := range ids {
		if _, err := metaStation.addItemToStation(ctx, id, channelUsername, channelFeedUrl); err != nil {
			log.Printf("error downloading video %s, error: %v\n", id, err)
		}
	}
	return metaStation.updateFeed()
}

func (metaStation *MetaStation) deleteVideo(videoTitle, author string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var index int = -1
	for i, item := range metaStation.Items {
		if item.Title == videoTitle && item.ITunesAuthor == author {
			index = i
			break
		}
	}
	if index == -1 {
		return errors.New("video does not exist in this show")
	}
	if err := Megh.deleteEpisode(ctx, metaStation.Items[index].GUID, metaStation.Title); err != nil {
		return err
	}
	metaStation.Items = append(metaStation.Items[:index], metaStation.Items[index+1:]...)
	metaStation.updateFeed()
	// fmt.Printf("file deleted with id %v\n", id)
	return nil
}

func (metaStation *MetaStation) delete() error {
	if err := Megh.deleteShow(metaStation.Title); err != nil {
		return err
	}
	os.Remove(Megh.getLocalFeedFilepath(metaStation.Title))
	os.Remove(Megh.getLocalStationFilepath(metaStation.Title))
	StationNames.Remove(metaStation.Title)
	return nil
}

func (metaStation *MetaStation) addVideo(videoUrl string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	id, err := getVideoId(ctx, videoUrl)
	if err != nil {
		return "", err
	}
	ids := metaStation.filter([]string{id})
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
	if share, err := metaStation.addItemToStation(ctx, id, username, channelFeedUrl); err != nil {
		return "", err
	} else {
		return share, nil
	}
}

func (metaStation *MetaStation) addItemToStation(ctx context.Context, id, username, channelFeedUrl string) (string, error) {
	metaStationItem := MetaStationItem{
		GUID:           id,
		ITunesAuthor:   username,
		ChannelID:      channelFeedUrl,
		AddedOn:        time.Now(),
		ITunesExplicit: "no",
		Link:           "https://www.youtube.com/watch?v=" + id,
	}
	if err := isValidForDownload(ctx, metaStationItem.Link); err != nil {
		return "", err
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
			// fmt.Printf("Error: %v\n", err)
			return
		} else {
			description = "Link to the YouTube Video: " + metaStationItem.Link + "\n" + description
			metaStationItem.Description = description
			metaStationItem.ITunesSubtitle = description
		}
	}()
	wg.Add(1)
	go func() {
		wg.Done()
		if duration, err := getVideoDuration(ctx, metaStationItem.Link); err != nil {
			// fmt.Printf("Error: %v\n", err)
			return
		} else {
			metaStationItem.ITunesDuration = duration
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		if views, err := getVideoViews(ctx, metaStationItem.Link); err != nil {
			// fmt.Printf("Error: %v\n", err)
			return
		} else {
			metaStationItem.Views = views
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		if pubDate, err := getVideoPubDate(ctx, metaStationItem.Link); err != nil {
			// fmt.Printf("Error: %v\n", err)
			return
		} else {
			metaStationItem.PubDate = pubDate
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := metaStationItem.saveVideoThumbnail(ctx, metaStation.Title, metaStationItem.Link); err != nil {
			// fmt.Printf("Error: %v\n", err)
			return
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		if size, err := metaStationItem.saveAudio(ctx, metaStation.Title, metaStationItem.Link, 0); err != nil {
			// fmt.Printf("Error: %v\n", err)
			return
		} else {
			metaStation.makeSpace(ctx, size)
			if share, err := Megh.upload(ctx, metaStationItem.GUID, metaStation.Title, AUDIO); err != nil {
				// fmt.Printf("audio error: %v\n", err)
				return
			} else {
				metaStationItem.Enclosure = Enclosure{
					URL:    share,
					Type:   "audio/mpeg",
					Length: size,
				}
				if share, err := Megh.upload(ctx, metaStationItem.GUID, metaStation.Title, THUMBNAIL); err != nil {
					// fmt.Printf("thumbnail error: %v\n", err)
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
	return metaStation.updateFeed()
}

func (metaStation *MetaStation) getLatestVideos(ctx context.Context, channelUrl string, limit uint) ([]string, error) {
	args := []string{
		"--get-id",
		"--match-filter",
		"live_status!='is_upcoming'",
		"--playlist-end",
		fmt.Sprint(limit),
		channelUrl,
	}
	out, err := run(ctx, "yt-dlp", args...)
	if err != nil {
		if strings.Contains(err.Error(), "Premieres in") {
			return nil, nil
		}
		return nil, err
	}

	ids := strings.Split(strings.TrimSpace(out), "\n")
	videoIds := metaStation.filter(ids)
	return videoIds, nil
}

func getVideoId(ctx context.Context, link string) (string, error) {
	out, err := run(
		ctx,
		"yt-dlp",
		"--quiet",
		"--print",
		"id",
		link,
	)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func getVideoUsername(ctx context.Context, link string) (string, error) {
	out, err := run(
		ctx,
		"yt-dlp",
		"--quiet",
		"--print",
		"uploader_id",
		link,
	)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func getVideoTitle(ctx context.Context, link string) (string, error) {
	out, err := run(
		ctx,
		"yt-dlp",
		"--quiet",
		"--print",
		"title",
		link,
	)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func getVideoDescription(ctx context.Context, link string) (string, error) {
	out, err := run(
		ctx,
		"yt-dlp",
		"--quiet",
		"--print",
		"description",
		link,
	)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func getVideoDuration(ctx context.Context, link string) (string, error) {
	out, err := run(
		ctx,
		"yt-dlp",
		"--quiet",
		"--print",
		"duration_string",
		link,
	)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func isValidForDownload(ctx context.Context, link string) error {
	out, err := run(
		ctx,
		"yt-dlp",
		"--quiet",
		"--print",
		"duration",
		link,
	)
	if err != nil {
		return err
	}

	durationSeconds, err := strconv.Atoi(strings.TrimSpace(out))
	if err != nil {
		return err
	}
	fiveHoursInSeconds := 5 * 60 * 60
	if durationSeconds > fiveHoursInSeconds {
		return errors.New("video needs to be shorter than 5 hours long!")
	}
	return nil
}

func getVideoViews(ctx context.Context, link string) (uint32, error) {
	out, err := run(
		ctx,
		"yt-dlp",
		"--quiet",
		"--print",
		"view_count",
		link,
	)
	if err != nil {
		return 0, err
	}
	n, err := strconv.Atoi(strings.TrimSpace(out))
	if err != nil {
		return 0, err
	}
	return uint32(n), nil
}

func getVideoPubDate(ctx context.Context, link string) (string, error) {
	out, err := run(
		ctx,
		"yt-dlp",
		"--quiet",
		"--print",
		"upload_date",
		link,
	)
	if err != nil {
		return "", err
	}
	out = strings.TrimSpace(out)
	return formatDate(out)
}

func (metaStationItem *MetaStationItem) saveVideoThumbnail(ctx context.Context, title string, link string) error {
	if err := os.MkdirAll(THUMBNAIL_BASE, 0o755); err != nil {
		return err
	}
	localpath2 := Megh.getLocalThumbnailFilepath(metaStationItem.GUID, title)
	localpath1 := strings.Split(localpath2, ".")[0] + ".webp"
	_, err := run(
		ctx,
		"yt-dlp",
		"--quiet",
		"--skip-download",
		"--write-thumbnail",
		"-o",
		strings.Split(localpath2, ".")[0]+".%(ext)s",
		link,
	)
	if err == nil {
		// fmt.Printf("thumbnail downloaded\n")

	}
	ConvertImageToCorrectFormat(localpath1, localpath2)
	return err
}

func (metaStationItem *MetaStationItem) saveAudio(ctx context.Context, title string, link string, audioQuality int) (uint64, error) {
	if err := os.MkdirAll(AUDIO_BASE, 0o755); err != nil {
		return 0, err
	}
	localpath := Megh.getLocalAudioFilepath(metaStationItem.GUID, title)
	_, err := run(ctx,
		"yt-dlp",
		"--quiet",
		"-x",
		"--audio-format",
		"mp3",
		"--audio-quality",
		strconv.Itoa(audioQuality),
		"-o",
		strings.Split(localpath, ".")[0]+".%(ext)s",
		link,
	)
	if err != nil {
		return 0, err
	}
	if err == nil {
		// fmt.Printf("audio downloaded\n")

	}
	if info, err := os.Stat(localpath); err != nil {
		return 0, err
	} else {
		return uint64(info.Size()), nil
	}
}

func formatDate(uploadDate string) (string, error) {
	t, err := time.Parse("20060102", uploadDate)
	if err != nil {
		return "", err
	}
	return t.Format("Mon, 02 Jan 2006 15:04:05 GMT"), nil
}

func (metaStation *MetaStation) updateFeed() (string, error) {
	station := metaStation.getStation()
	metaStation.saveMetaStationToLocal()
	return station.saveFeed()
}

func (metaStation *MetaStation) getStation() Station {
	return Station{
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
}

func GetChannelFeedUrl(username string) (string, error) {
	if len(username) == 0 {
		return "", fmt.Errorf("username is empty")
	}
	if username[0] != '@' {
		username = "@" + username
	}
	channelFeedUrl := "https://www.youtube.com/" + username + "/videos"
	if isValidUrl(channelFeedUrl) == false {
		return "", fmt.Errorf("not valid username\n")
	}
	return channelFeedUrl, nil
}

func isValidUrl(url string) bool {
	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}
	return true
}
