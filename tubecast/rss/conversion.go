package rss

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/sharing"
)

func getLatestVideos(ctx context.Context, channelUrl string, limit uint) ([]string, error) {
	out, err := run(
		ctx,
		"yt-dlp",
		"--get-id",
		"--playlist-end",
		fmt.Sprint(limit),
		channelUrl,
	)
	if err != nil {
		return nil, err
	}
	ids := strings.Split(strings.TrimSpace(out), "\n")
	fmt.Printf("%#v\n", ids)
	return ids, nil
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

func saveVideoThumbnail(ctx context.Context, videoID string) error {
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
		THUMBNAILS_BASE+"/%(id)s.%(ext)s",
		"https://www.youtube.com/watch?v="+videoID,
	)
	return err
}

func saveAudio(ctx context.Context, videoID string) error {
	if err := os.MkdirAll(AUDIO_BASE, 0o755); err != nil {
		return err
	}
	_, err := run(ctx,
		"yt-dlp",
		"--quiet",
		"-x",
		"--audio-format",
		"mp3",
		"--audio-quality",
		"0",
		"-o", filepath.Join(AUDIO_BASE, "%(id)s.%(ext)s"),
		"https://www.youtube.com/watch?v="+videoID,
	)
	return err
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
func uploadToDropbox(localPath, dropboxPath string) (string, error) {
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

	existingFiles, err := listFileNames(accessToken, DROPBOX_AUDIO_BASE)
	if err != nil {
		return "", fmt.Errorf("failed to list files: %w", err)
	}
	fmt.Println("Files:", existingFiles)

	file := path.Base(dropboxPath)
	if !slices.Contains(existingFiles, file) {
		fmt.Printf("uploading %v\n", dropboxPath)
		uploadArg := files.NewUploadArg(dropboxPath)
		uploadArg.Mode.Tag = "overwrite"
		uploadArg.Mute = true
		_, err = dbx.Upload(uploadArg, f)
		if err != nil {
			return "", fmt.Errorf("dropbox upload: %w", err)
		}
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
		shareURL = shareURL + "?raw=1"
	}

	return shareURL, nil
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

func Init() {
	TOKEN_MANAGER = NewTokenManager(
		os.Getenv("DROPBOX_APP_KEY"),
		os.Getenv("DROPBOX_APP_SECRET"),
		os.Getenv("DROPBOX_REFRESH_TOKEN"),
	)
}
