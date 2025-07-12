package rss

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Atomatically save Station Meta data locally
func (metaStation *MetaStation) saveMetaStationToLocal() error {
	path := Megh.getLocalStationFilepath(metaStation.Title)
	tmp := path + ".tmp"
	if err := os.MkdirAll(filepath.Dir(tmp), 0o755); err != nil {
		return err
	}
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	defer f.Close()
	defer os.Remove(tmp)
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(metaStation); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Loads Station Meta data from the local
func loadMetaStationFromLocal(path string) (MetaStation, error) {
	f, err := os.Open(path)
	if err != nil {
		return MetaStation{}, err
	}
	defer f.Close()
	var metaStation MetaStation
	dec := json.NewDecoder(f)
	if err := dec.Decode(&metaStation); err != nil {
		return MetaStation{}, err
	}
	// if metaStation.SubscribedChannel == nil {
	// 	metaStation.SubscribedChannel = NewSet[string]()
	// 	return metaStation, errors.New("black magic")
	// }
	return metaStation, nil
}

// Atomatically save Station data locally
func (station *Station) saveXMLToLocal() (string, error) {
	if err := os.MkdirAll(FEED_BASE, 0o755); err != nil {
		return "", err
	}
	path := Megh.getLocalFeedFilepath(station.Title)
	tmp := path + ".tmp"

	f, err := os.Create(tmp)
	if err != nil {
		return "", err
	}

	defer f.Close()
	defer os.Remove(tmp)

	enc := xml.NewEncoder(f)
	enc.Indent("", "  ")
	f.WriteString(xml.Header)
	f.WriteString("<rss xmlns:itunes=\"http://www.itunes.com/dtds/podcast-1.0.dtd\" xmlns:podcast=\"https://podcastindex.org/namespace/1.0\" version=\"2.0\">\n  ")
	// f.WriteString("<rss xmlns:itunes=\"http://www.itunes.com/dtds/podcast-1.0.dtd\" version=\"2.0\">")
	if err := enc.Encode(station); err != nil {
		return "", err
	}
	f.WriteString("\n</rss>")
	if err = os.Rename(tmp, path); err != nil {
		return "", err
	}

	if Megh.IsArchive {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()
		return Megh.upload(ctx, "", station.Title, FEED)
	}
	return Megh.getShareableFeedUrl(station.Title), nil
}

func loadAllMetaStationNames() error {
	StationNames = NewSet[string]()
	entries, err := os.ReadDir(STATION_BASE)
	if err != nil {
		return nil
	}
	for _, e := range entries {
		if !e.IsDir() {
			StationNames.Add(strings.TrimSuffix(e.Name(), filepath.Ext(e.Name())))
		}
	}
	// for title := range StationNames.Value {
	// 	fmt.Printf("Station %v\n", title)
	// }
	return nil
}

func (cloud *Cloud) upload(ctx context.Context, id, title string, filetype FileType) (string, error) {
	var localpath, remotepath string
	var isLocalDelete bool
	switch filetype {
	case THUMBNAIL:
		localpath = cloud.getLocalThumbnailFilepath(id, title)
		remotepath = cloud.getShareableThumbnailUrl(id, title)
		isLocalDelete = true
	case AUDIO:
		localpath = cloud.getLocalAudioFilepath(id, title)
		remotepath = cloud.getShareableAudioUrl(id, title)
		isLocalDelete = true
	case FEED:
		localpath = cloud.getLocalFeedFilepath(title)
		remotepath = cloud.getShareableFeedUrl(title)
	case COVER:
		localpath = cloud.getLocalCoverFilepath(title)
		remotepath = cloud.getShareableCoverUrl(title)
	default:
		return "", errors.New("wrong filetype being uploaded")
	}
	fmt.Printf("localpath: %v\n", localpath)
	if _, err := os.Open(localpath); err != nil {
		fmt.Printf("could not open localpath: %v\n", localpath)
		return "", err
	}
	_, err := run(
		ctx,
		"ia",
		"upload",
		"--no-backup",
		cloud.ArchiveId,
		localpath,
	)
	if err != nil {
		return "", err
	}
	if isLocalDelete {
		os.Remove(localpath)
	}
	fmt.Printf("file uploaded!\n")
	return remotepath, nil
}

func (cloud *Cloud) getUsage(ctx context.Context) (Usage, error) {
	out, err := run(
		ctx,
		"ia",
		"metadata",
		cloud.ArchiveId,
	)
	if err != nil {
		return Usage{}, err
	}

	// Include the file name so we can filter out history files
	var meta struct {
		Files []struct {
			Name string `json:"name"`
			Size any    `json:"size"`
		} `json:"files"`
	}
	if err := json.Unmarshal([]byte(out), &meta); err != nil {
		return Usage{}, err
	}

	var totalBytes uint64
	var count uint64
	for _, f := range meta.Files {
		if strings.HasPrefix(f.Name, "history/") {
			continue
		}
		switch v := f.Size.(type) {
		case float64:
			totalBytes += uint64(v)
		case string:
			if n, err := strconv.ParseUint(v, 10, 64); err == nil {
				totalBytes += n
			}
		}
		count++
	}

	fmt.Printf("total size used (excluding history): %v MB, total files: %v\n", totalBytes/(1024*1024), count)
	return Usage{
		TotalSizeBytes: totalBytes,
		TotalSizeMiB:   totalBytes / (1024 * 1024),
		FileCount:      count,
	}, nil
}

func (cloud *Cloud) deleteEpisode(ctx context.Context, id, title string) error {
	_, err := run(
		ctx,
		"ia",
		"delete",
		cloud.ArchiveId,
		fmt.Sprintf("--glob=*%s_%s*", title, id),
	)
	if err != nil {
		fmt.Printf("delete error-1: %v\n", err)
	}
	// _, err = run(
	// 	ctx,
	// 	"./ia",
	// 	"delete",
	// 	cloud.ArchiveId,
	// 	fmt.Sprintf("--glob=***%s_%s*", title, id),
	// )
	// if err != nil {
	// 	fmt.Printf("delete error-2: %v\n", err)
	// }
	return err
}

func (cloud *Cloud) deleteShow(title string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	_, err := run(
		ctx,
		"ia",
		"delete",
		cloud.ArchiveId,
		fmt.Sprintf("--glob=*_%s_*", title),
	)
	if err != nil {
		return err
	}
	_, err = run(
		ctx,
		"ia",
		"delete",
		cloud.ArchiveId,
		fmt.Sprintf("--glob=cover_%s*", title),
	)
	_, err = run(
		ctx,
		"ia",
		"delete",
		cloud.ArchiveId,
		fmt.Sprintf("--glob=*%s.xml", title),
	)
	return err
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
