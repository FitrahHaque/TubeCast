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
	fmt.Println("Latest video urls to be uploaded: ", ids)
	for _, id := range ids {
		if _, err := metaStation.addItemToStation(ctx, id, channelUsername, channelFeedUrl); err != nil {
			log.Printf("error downloading video %s, error: %v\n", id, err)
		}
	}
	return metaStation.updateFeed()
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
		if err := metaStationItem.saveVideoThumbnail(ctx, metaStation.Title, metaStationItem.Link); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		if size, err := metaStationItem.saveAudio(ctx, metaStation.Title, metaStationItem.Link, 0); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		} else {
			metaStation.makeSpace(ctx, size)
			if share, err := Megh.upload(ctx, metaStationItem.GUID, metaStation.Title, AUDIO); err != nil {
				return
			} else {
				metaStationItem.Enclosure = Enclosure{
					URL:    share,
					Type:   "audio/mpeg",
					Length: size,
				}
				if share, err := Megh.upload(ctx, metaStationItem.GUID, metaStation.Title, THUMBNAIL); err != nil {
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
		"--cookies",
		"cookies.txt",
		"--ffmpeg-location",
		FFMPEG_PATH,
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
		"--cookies",
		"cookies.txt",
		"--ffmpeg-location",
		FFMPEG_PATH,
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
		"--cookies",
		"cookies.txt",
		"--ffmpeg-location",
		FFMPEG_PATH,
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
		"--cookies",
		"cookies.txt",
		"--ffmpeg-location",
		FFMPEG_PATH,
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
		"--cookies",
		"cookies.txt",
		"--ffmpeg-location",
		FFMPEG_PATH,
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
		"--cookies",
		"cookies.txt",
		"--ffmpeg-location",
		FFMPEG_PATH,
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

func getVideoViews(ctx context.Context, link string) (uint32, error) {
	out, err := run(
		ctx,
		"yt-dlp",
		"--cookies",
		"cookies.txt",
		"--ffmpeg-location",
		FFMPEG_PATH,
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
		"--cookies",
		"cookies.txt",
		"--ffmpeg-location",
		FFMPEG_PATH,
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
		"--cookies",
		"cookies.txt",
		"--ffmpeg-location",
		FFMPEG_PATH,
		"--quiet",
		"--skip-download",
		"--write-thumbnail",
		"-o",
		strings.Split(localpath2, ".")[0]+".%(ext)s",
		link,
	)
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
		"--cookies",
		"cookies.txt",
		"--ffmpeg-location",
		FFMPEG_PATH,
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
	return station.saveXMLToLocal()
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

// func listFileNames(token, path string) ([]string, error) {
// 	cfg := dropbox.Config{Token: token}
// 	dbx := files.New(cfg)

// 	arg := files.NewListFolderArg(path)
// 	arg.Recursive = false
// 	res, err := dbx.ListFolder(arg)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var names []string
// 	for {
// 		for _, e := range res.Entries {
// 			if f, ok := e.(*files.FileMetadata); ok {
// 				names = append(names, f.Name)
// 			}
// 		}
// 		if !res.HasMore {
// 			break
// 		}
// 		res, err = dbx.ListFolderContinue(
// 			&files.ListFolderContinueArg{Cursor: res.Cursor})
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
// 	return names, nil
// }

// dirSize calculates the total size of files in the given local directory path recursively.
// func dirSize(path string) (uint64, error) {
// 	var total uint64 = 0
// 	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
// 		if err != nil {
// 			return err
// 		}
// 		if !info.IsDir() {
// 			total += uint64(info.Size())
// 		}
// 		return nil
// 	})
// 	if err != nil {
// 		return 0, err
// 	}
// 	return total, nil
// }

// // UploadToDropbox uploads localPath into your app folder under dropboxPath
// // and returns the shared link URL
// func (station *Station) uploadItemMediaToDropbox(fileType FileType, id string) (string, error) {
// 	if stationItem, ok := station.GetStationItem(id); ok {
// 		switch fileType {
// 		case THUMBNAIL:
// 			return stationItem.ITunesImage.Href, nil
// 		case AUDIO:
// 			return stationItem.Enclosure.URL, nil
// 		}
// 	}
// 	var localPath, dropboxPath string
// 	switch fileType {
// 	case THUMBNAIL:
// 		localPath = filepath.Join(THUMBNAILS_BASE, station.Title, id+".png")
// 		dropboxPath = filepath.Join(DROPBOX_THUMBNAILS_BASE, station.Title, id+".png")
// 	case AUDIO:
// 		localPath = filepath.Join(AUDIO_BASE, station.Title, id+".mp3")
// 		dropboxPath = filepath.Join(DROPBOX_AUDIO_BASE, station.Title, id+".mp3")
// 	}

// 	if share, err := dropboxUpload(localPath, dropboxPath, true); err != nil {
// 		return "", err
// 	} else {
// 		return share, nil
// 	}
// }

// func dropboxUpload(localPath, dropboxPath string, deleteLocal bool) (string, error) {
// 	f, err := os.Open(localPath)
// 	if err != nil {
// 		return "", fmt.Errorf("open file: %w", err)
// 	}
// 	defer f.Close()

// 	accessToken, err := TOKEN_MANAGER.GetValidAccessToken()
// 	if err != nil {
// 		return "", fmt.Errorf("failed to get valid access token: %w", err)
// 	}

// 	cfg := dropbox.Config{Token: accessToken}
// 	dbx := files.New(cfg)
// 	// fmt.Printf("used: %vMiB\n", used/1024/1024)

// 	fmt.Printf("uploading %v\n", dropboxPath)
// 	uploadArg := files.NewUploadArg(dropboxPath)
// 	uploadArg.Mode.Tag = "overwrite"
// 	uploadArg.Mute = true
// 	_, err = dbx.Upload(uploadArg, f)
// 	if err != nil {
// 		return "", fmt.Errorf("dropbox upload: %w", err)
// 	}

// 	sharingClient := sharing.New(cfg)

// 	shareArg := &sharing.CreateSharedLinkWithSettingsArg{
// 		Path:     dropboxPath,
// 		Settings: nil,
// 	}

// 	shareLink, err := sharingClient.CreateSharedLinkWithSettings(shareArg)
// 	if err != nil {
// 		if strings.Contains(err.Error(), "shared_link_already_exists") {
// 			listArg := &sharing.ListSharedLinksArg{
// 				Path:       dropboxPath,
// 				DirectOnly: true,
// 			}

// 			listResult, listErr := sharingClient.ListSharedLinks(listArg)
// 			if listErr != nil {
// 				return "", fmt.Errorf("list shared links: %w", listErr)
// 			}

// 			if len(listResult.Links) == 0 {
// 				return "", fmt.Errorf("no shared links found for path: %s", dropboxPath)
// 			}

// 			shareLink = listResult.Links[0]
// 		} else {
// 			return "", fmt.Errorf("create shared link: %w", err)
// 		}
// 	}

// 	var shareURL string
// 	switch link := shareLink.(type) {
// 	case *sharing.FileLinkMetadata:
// 		shareURL = link.Url
// 	case *sharing.FolderLinkMetadata:
// 		shareURL = link.Url
// 	default:
// 		return "", fmt.Errorf("unexpected link type")
// 	}

// 	if len(shareURL) > 0 {
// 		// u := strings.ReplaceAll(shareURL, "\u0026", "&")
// 		shareURL = strings.Split(shareURL, "dl=0")[0] + "raw=1"
// 		if shareURL, err = fetchFinalURL(shareURL); err != nil {
// 			return shareURL, err
// 		}
// 		if deleteLocal {
// 			os.Remove(localPath)
// 		}
// 		return shareURL, nil
// 	}
// 	return shareURL, fmt.Errorf("empty url\n")
// }

// // NewTokenManager creates a new token manager
// func NewTokenManager(appKey, appSecret, refreshToken string) *TokenManager {
// 	return &TokenManager{
// 		AppKey:       appKey,
// 		AppSecret:    appSecret,
// 		RefreshToken: refreshToken,
// 	}
// }

// // RefreshAccessToken gets a new access token using the refresh token
// func (tm *TokenManager) RefreshAccessToken() error {
// 	data := url.Values{}
// 	data.Set("grant_type", "refresh_token")
// 	data.Set("refresh_token", tm.RefreshToken)

// 	req, err := http.NewRequest("POST", "https://api.dropbox.com/oauth2/token", strings.NewReader(data.Encode()))
// 	if err != nil {
// 		return fmt.Errorf("failed to create request: %w", err)
// 	}

// 	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
// 	req.SetBasicAuth(tm.AppKey, tm.AppSecret)

// 	client := &http.Client{Timeout: 30 * time.Second}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return fmt.Errorf("failed to refresh token: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return fmt.Errorf("failed to read response: %w", err)
// 	}

// 	if resp.StatusCode != http.StatusOK {
// 		return fmt.Errorf("refresh token failed: %s - %s", resp.Status, string(body))
// 	}

// 	var tokenResp DropboxTokenResponse
// 	if err := json.Unmarshal(body, &tokenResp); err != nil {
// 		return fmt.Errorf("failed to parse token response: %w", err)
// 	}

// 	tm.AccessToken = tokenResp.AccessToken
// 	tm.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

// 	return nil
// }

// GetValidAccessToken returns a valid access token, refreshing if necessary
// func (tm *TokenManager) GetValidAccessToken() (string, error) {
// 	if tm.AccessToken == "" || time.Now().Add(5*time.Minute).After(tm.ExpiresAt) {
// 		if err := tm.RefreshAccessToken(); err != nil {
// 			return "", err
// 		}
// 	}
// 	return tm.AccessToken, nil
// }

// func deleteDropboxFile(accessToken, dropboxPath string) error {
// 	cfg := dropbox.Config{Token: accessToken}
// 	dbx := files.New(cfg)

// 	arg := &files.DeleteArg{Path: dropboxPath}

// 	_, err := dbx.DeleteV2(arg)
// 	if err != nil {
// 		return fmt.Errorf("failed to delete %q: %w", dropboxPath, err)
// 	}

// 	fmt.Printf("Deleted file at %s\n", dropboxPath)
// 	return nil
// }
