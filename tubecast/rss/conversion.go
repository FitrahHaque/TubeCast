package rss

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/sharing"
)

func getChannelVideoIDs(ctx context.Context, channelUrl string, limit uint) ([]string, error) {
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
		"\"%(title)s\"",
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
		"\"%(description)s\"",
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
		"\"%(duration_string)s\"",
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
		"\"%(view_count)s\"",
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
		"\"%(upload_date)s\"",
		"https://www.youtube.com/watch?v="+videoID,
	)
	if err != nil {
		return "", err
	}
	return formatDate(strings.TrimSpace(out))
}

func saveVideoThumbnail(ctx context.Context, videoID string) error {
	if err := os.MkdirAll("../thumbnails", 0o755); err != nil {
		return err
	}
	_, err := run(
		ctx,
		"yt-dlp",
		"--quiet",
		"--skip-download",
		"--write-thumbnail",
		"-o", "../thumbnails/%(id)s.%(ext)s",
		"https://www.youtube.com/watch?v="+videoID,
	)
	return err
}

func saveAudio(ctx context.Context, videoID string) error {
	if err := os.MkdirAll("../audio", 0o755); err != nil {
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
		"-o", filepath.Join("../audio", "%(id)s.%(ext)s"),
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

// uploadToDropbox uploads localPath (e.g. "audio/vDWaKVmqznQ.mp3")
// into your app folder under dropboxPath (e.g. "/PodcastAudio/vDWaKVmqznQ.mp3")
// and returns the shared link URL.
func (user *User) UploadToDropbox(localPath, dropboxPath string) (string, error) {
	// 1) Open the file
	f, err := os.Open(localPath)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	// 2) Configure the client
	cfg := dropbox.Config{Token: os.Getenv("DROPBOX_TOKEN")}
	dbx := files.New(cfg)

	// 3) Upload
	uploadArg := files.NewUploadArg(dropboxPath)
	uploadArg.Mode.Tag = "overwrite" // overwrite if it exists
	uploadArg.Mute = true            // no notifications
	_, err = dbx.Upload(uploadArg, f)
	if err != nil {
		return "", fmt.Errorf("dropbox upload: %w", err)
	}

	// 4) Create a shared link (if you want a public URL)
	// Create a sharing client using the same config
	sharingClient := sharing.New(cfg)

	// Create the argument for CreateSharedLinkWithSettings
	shareArg := &sharing.CreateSharedLinkWithSettingsArg{
		Path: dropboxPath,
		// Settings can be nil for default settings
		Settings: nil,
	}

	shareLink, err := sharingClient.CreateSharedLinkWithSettings(shareArg)
	if err != nil {
		return "", fmt.Errorf("create shared link: %w", err)
	}

	// Extract the URL from the response
	var shareURL string
	switch link := shareLink.(type) {
	case *sharing.FileLinkMetadata:
		shareURL = link.Url
	case *sharing.FolderLinkMetadata:
		shareURL = link.Url
	default:
		return "", fmt.Errorf("unexpected link type")
	}

	// By default Dropbox shared links point at a download page.
	// To get a direct file link, replace ?dl=0 with ?raw=1:
	// This might not always work depending on the link format
	if len(shareURL) > 0 {
		// Simple URL modification for direct download
		return shareURL + "?raw=1", nil
	}

	return shareURL, nil
}
