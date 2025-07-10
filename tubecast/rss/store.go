package rss

import (
	"context"
	"encoding/json"
	"encoding/xml"
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
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
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
	if os.IsNotExist(err) {
		return MetaStation{}, nil
	}
	if err != nil {
		return MetaStation{}, err
	}
	defer f.Close()
	var station MetaStation
	dec := json.NewDecoder(f)
	if err := dec.Decode(&station); err != nil {
		return MetaStation{}, err
	}
	return station, nil
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
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	return Megh.upload(ctx, "", station.Title, FEED)
}

func loadAllMetaStationNames() error {
	entries, err := os.ReadDir(STATION_BASE)
	if err != nil {
		return err
	}
	StationNames = NewSet[string]()
	for _, e := range entries {
		if !e.IsDir() {
			StationNames.Add(strings.TrimSuffix(e.Name(), filepath.Ext(e.Name())))
		}
	}
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
	}
	fmt.Printf("localpath: %v\n", localpath)
	if _, err := os.Open(localpath); err != nil {
		return "", err
	}
	_, err := run(
		ctx,
		"./ia",
		"upload",
		cloud.ArchiveId,
		localpath,
	)
	if err != nil {
		return "", err
	}
	if isLocalDelete {
		os.Remove(localpath)
	}
	return fetchFinalURL(remotepath)
}

func (cloud *Cloud) getUsage(ctx context.Context) (Usage, error) {
	out, err := run(
		ctx,
		"./ia",
		"metadata",
		cloud.ArchiveId,
	)
	if err != nil {
		return Usage{}, err
	}

	var meta struct {
		Files []struct {
			Size interface{} `json:"size"`
		} `json:"files"`
	}

	if err := json.Unmarshal([]byte(out), &meta); err != nil {
		return Usage{}, err
	}

	var totalBytes uint64
	for _, f := range meta.Files {
		switch v := f.Size.(type) {
		case float64:
			totalBytes += uint64(v)
		case string:
			if n, err := strconv.ParseUint(v, 10, 64); err == nil {
				totalBytes += n
			}
		}
	}
	fileCount := uint64(len(meta.Files))
	fmt.Printf("totalbytes used: %v, total files: %v\n", totalBytes, fileCount)
	return Usage{
		TotalSizeBytes: totalBytes,
		TotalSizeMiB:   totalBytes / (1024 * 1024),
		FileCount:      fileCount,
	}, nil
}

func (cloud *Cloud) delete(ctx context.Context, id string, filename string) error {
	_, err := run(
		ctx,
		"./ia",
		"remove",
		cloud.ArchiveId,
		filename,
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
