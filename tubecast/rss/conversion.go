package rss

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
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

func getVideoThumbnail(ctx context.Context, videoID string) error {
	out1, err := run(
		ctx,
		"yt-dlp",
		"--quiet",
		"--print",
		"\"%(thumbnail)s\"",
		"--skip-download",
		"--write-thumbnail",
		"https://www.youtube.com/watch?v="+videoID,
	)
	if err != nil {
		return err
	}
	out1 = strings.TrimSpace(out1)
	_, err = run(
		ctx,
		"curl",
		"-o",
		"thumbnail-"+videoID,
		out1,
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
