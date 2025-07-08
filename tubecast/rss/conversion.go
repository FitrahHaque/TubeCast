package rss

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/sharing"
)

func (station *Station) getLatestVideos(ctx context.Context, channelUrl string, limit uint) ([]string, error) {
	// Build the arguments so that each option gets the right argument:
	args := []string{
		"--get-id",
		"--match-filter", "live_status!='is_upcoming'",
		"--playlist-end", fmt.Sprint(limit),
		channelUrl,
	}

	out, err := run(ctx, "yt-dlp", args...)
	if err != nil {
		// Handle the “Premieres in…” message if needed:
		if strings.Contains(err.Error(), "Premieres in") {
			return nil, nil
		}
		return nil, err
	}

	ids := strings.Split(strings.TrimSpace(out), "\n")
	var videoIds []string
	for _, id := range ids {
		if !station.HasItem(id) {
			videoIds = append(videoIds, id)
		}
	}
	return videoIds, nil
}

func getVideoTitle(ctx context.Context, videoID string) (string, error) {
	out, err := run(
		ctx,
		"yt-dlp",
		"--quiet",
		"--print",
		"%(title)s",
		"https://www.youtube.com/watch?v="+videoID,
	)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func getVideoDescription(ctx context.Context, videoID string) (string, error) {
	out, err := run(
		ctx,
		"yt-dlp",
		"--quiet",
		"--print",
		"%(description)s",
		"https://www.youtube.com/watch?v="+videoID,
	)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func getVideoDuration(ctx context.Context, videoID string) (string, error) {
	out, err := run(
		ctx,
		"yt-dlp",
		"--quiet",
		"--print",
		"%(duration_string)s",
		"https://www.youtube.com/watch?v="+videoID,
	)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func getVideoViews(ctx context.Context, videoID string) (uint32, error) {
	out, err := run(
		ctx,
		"yt-dlp",
		"--quiet",
		"--print",
		"%(view_count)s",
		"https://www.youtube.com/watch?v="+videoID,
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

func getVideoPubDate(ctx context.Context, videoID string) (string, error) {
	out, err := run(
		ctx,
		"yt-dlp",
		"--quiet",
		"--print",
		"%(upload_date)s",
		"https://www.youtube.com/watch?v="+videoID,
	)
	if err != nil {
		return "", err
	}
	out = strings.TrimSpace(out)
	// fmt.Printf("Date: %v\n", out)
	return formatDate(out)
}

func (metaStationItem *MetaStationItem) saveVideoThumbnail(ctx context.Context, videoID string) error {
	if err := os.MkdirAll(THUMBNAILS_BASE, 0o755); err != nil {
		return err
	}
	_, err := run(
		ctx,
		"yt-dlp",
		"--quiet",
		"--skip-download",
		"--write-thumbnail",
		"-o",
		THUMBNAILS_BASE+"/"+metaStationItem.ID+".%(ext)s",
		"https://www.youtube.com/watch?v="+videoID,
	)
	ConvertImageToCorrectFormat(THUMBNAILS_BASE, metaStationItem.ID)
	return err
}

func (metaStationItem *MetaStationItem) saveAudio(ctx context.Context, videoID string) (uint64, error) {
	if err := os.MkdirAll(AUDIO_BASE, 0o755); err != nil {
		return 0, err
	}
	_, err := run(ctx,
		"yt-dlp",
		"--quiet",
		"-x",
		"--audio-format",
		"mp3",
		"--audio-quality",
		"0",
		"-o",
		AUDIO_BASE+"/"+metaStationItem.ID+".%(ext)s",
		"https://www.youtube.com/watch?v="+videoID,
	)
	if err != nil {
		return 0, err
	}
	path := filepath.Join(AUDIO_BASE, metaStationItem.ID+".mp3")
	if info, err := os.Stat(path); err != nil {
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

func listFileNames(token, path string) ([]string, error) {
	cfg := dropbox.Config{Token: token}
	dbx := files.New(cfg)

	arg := files.NewListFolderArg(path)
	arg.Recursive = false
	res, err := dbx.ListFolder(arg)
	if err != nil {
		return nil, err
	}

	var names []string
	for {
		for _, e := range res.Entries {
			if f, ok := e.(*files.FileMetadata); ok {
				names = append(names, f.Name)
			}
		}
		if !res.HasMore {
			break
		}
		res, err = dbx.ListFolderContinue(
			&files.ListFolderContinueArg{Cursor: res.Cursor})
		if err != nil {
			return nil, err
		}
	}
	return names, nil
}

func dirSize(dbx files.Client, path string) (uint64, error) {
	arg := files.NewListFolderArg(path)
	arg.Recursive = true

	res, err := dbx.ListFolder(arg)
	if err != nil {
		return 0, err
	}
	var total uint64
	for {
		for _, entry := range res.Entries {
			if f, ok := entry.(*files.FileMetadata); ok {
				total += uint64(f.Size)
			}
		}
		if !res.HasMore {
			break
		}
		res, err = dbx.ListFolderContinue(&files.ListFolderContinueArg{Cursor: res.Cursor})
		if err != nil {
			return 0, err
		}
	}
	return total, nil
}

// UploadToDropbox uploads localPath into your app folder under dropboxPath
// and returns the shared link URL
func (station *Station) uploadItemMediaToDropbox(fileType FileType, id string) (string, error) {
	if stationItem, ok := station.GetStationItem(id); ok {
		switch fileType {
		case THUMBNAIL:
			return stationItem.ITunesImage.Href, nil
		case AUDIO:
			return stationItem.Enclosure.URL, nil
		}
	}
	var localPath, dropboxPath string
	switch fileType {
	case THUMBNAIL:
		localPath = filepath.Join(THUMBNAILS_BASE, id+".png")
		dropboxPath = filepath.Join(DROPBOX_THUMBNAILS_BASE, id+".png")
	case AUDIO:
		localPath = filepath.Join(AUDIO_BASE, id+".mp3")
		dropboxPath = filepath.Join(DROPBOX_AUDIO_BASE, id+".mp3")
	}

	if share, err := dropboxUpload(localPath, dropboxPath, true); err != nil {
		return "", err
	} else {
		return share, nil
	}
}

// fetchFinalURL follows redirects and returns the ultimate URL as a string.
func fetchFinalURL(rawURL string) (string, error) {
	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := client.Get(rawURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	return resp.Request.URL.String(), nil
}

func dropboxUpload(localPath, dropboxPath string, deleteLocal bool) (string, error) {
	f, err := os.Open(localPath)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	accessToken, err := TOKEN_MANAGER.GetValidAccessToken()
	if err != nil {
		return "", fmt.Errorf("failed to get valid access token: %w", err)
	}

	cfg := dropbox.Config{Token: accessToken}
	dbx := files.New(cfg)

	used, err := dirSize(dbx, DROPBOX_AUDIO_BASE)
	if err != nil {
		return "", err
	}
	stat, err := f.Stat()
	if err != nil {
		return "", err
	}
	fmt.Printf("used: %vMiB\n", used/1024/1024)
	if used+uint64(stat.Size()) > MaximumCloudStorage {
		return "", fmt.Errorf("quota exceeded\n")
	}

	fmt.Printf("uploading %v\n", dropboxPath)
	uploadArg := files.NewUploadArg(dropboxPath)
	uploadArg.Mode.Tag = "overwrite"
	uploadArg.Mute = true
	_, err = dbx.Upload(uploadArg, f)
	if err != nil {
		return "", fmt.Errorf("dropbox upload: %w", err)
	}

	sharingClient := sharing.New(cfg)

	shareArg := &sharing.CreateSharedLinkWithSettingsArg{
		Path:     dropboxPath,
		Settings: nil,
	}

	shareLink, err := sharingClient.CreateSharedLinkWithSettings(shareArg)
	if err != nil {
		if strings.Contains(err.Error(), "shared_link_already_exists") {
			listArg := &sharing.ListSharedLinksArg{
				Path:       dropboxPath,
				DirectOnly: true,
			}

			listResult, listErr := sharingClient.ListSharedLinks(listArg)
			if listErr != nil {
				return "", fmt.Errorf("list shared links: %w", listErr)
			}

			if len(listResult.Links) == 0 {
				return "", fmt.Errorf("no shared links found for path: %s", dropboxPath)
			}

			shareLink = listResult.Links[0]
		} else {
			return "", fmt.Errorf("create shared link: %w", err)
		}
	}

	var shareURL string
	switch link := shareLink.(type) {
	case *sharing.FileLinkMetadata:
		shareURL = link.Url
	case *sharing.FolderLinkMetadata:
		shareURL = link.Url
	default:
		return "", fmt.Errorf("unexpected link type")
	}

	if len(shareURL) > 0 {
		// u := strings.ReplaceAll(shareURL, "\u0026", "&")
		shareURL = strings.Split(shareURL, "dl=0")[0] + "raw=1"
		if shareURL, err = fetchFinalURL(shareURL); err != nil {
			return shareURL, err
		}
		if deleteLocal {
			os.Remove(localPath)
		}
		return shareURL, nil
	}
	return shareURL, fmt.Errorf("empty url\n")
}

func (station *Station) updateFeed() error {
	if err := station.saveXMLToLocal(); err != nil {
		return err
	}
	localPath := filepath.Join(FEED_BASE, station.Title+".xml")
	dropboxPath := filepath.Join(DROPBOX_FEED_BASE, station.Title+".xml")
	if share, err := dropboxUpload(localPath, dropboxPath, false); err != nil {
		return err
	} else {
		if metaStation, err := getMetaStation(station.Title); err != nil {
			return err
		} else {
			metaStation.Url = share
			station.Url = share
			fmt.Println("Feed updated")
			return metaStation.saveMetaStationToLocal()
		}
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

// NewTokenManager creates a new token manager
func NewTokenManager(appKey, appSecret, refreshToken string) *TokenManager {
	return &TokenManager{
		AppKey:       appKey,
		AppSecret:    appSecret,
		RefreshToken: refreshToken,
	}
}

// RefreshAccessToken gets a new access token using the refresh token
func (tm *TokenManager) RefreshAccessToken() error {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", tm.RefreshToken)

	req, err := http.NewRequest("POST", "https://api.dropbox.com/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(tm.AppKey, tm.AppSecret)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("refresh token failed: %s - %s", resp.Status, string(body))
	}

	var tokenResp DropboxTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return fmt.Errorf("failed to parse token response: %w", err)
	}

	tm.AccessToken = tokenResp.AccessToken
	tm.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return nil
}

// GetValidAccessToken returns a valid access token, refreshing if necessary
func (tm *TokenManager) GetValidAccessToken() (string, error) {
	if tm.AccessToken == "" || time.Now().Add(5*time.Minute).After(tm.ExpiresAt) {
		if err := tm.RefreshAccessToken(); err != nil {
			return "", err
		}
	}
	return tm.AccessToken, nil
}
