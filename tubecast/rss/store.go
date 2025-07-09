package rss

import (
	"encoding/json"
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
)

// Atomatically save Station Meta data locally
func (metaStation *MetaStation) saveMetaStationToLocal() error {
	path := filepath.Join(STATION_BASE, metaStation.Title+".json")
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
	enc.SetIndent("", " ")
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
	path := filepath.Join(FEED_BASE, station.Title+".xml")
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
	return path, nil
}

// Loads Station data from the local
// func loadXMLfromLocal() (Station, error) {
// 	f, err := os.Open(path)
// 	if os.IsNotExist(err) {
// 		return Station{}, nil
// 	}
// 	if err != nil {
// 		return Station{}, err
// 	}

// 	defer f.Close()

// 	var station Station
// 	dec := xml.NewDecoder(f)
// 	if err := dec.Decode(&station); err != nil {
// 		return Station{}, err
// 	}
// 	return station, nil
// }

// Save Station to cloud
// func UploadStation(station Station) error {

// }

// // Fetch Station from cloud
// func FetchStation(url string) (Station, error) {

// }
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

func getCoverImage(name string) (ITunesImage, error) {
	ConvertImageToCorrectFormat(COVER_BASE, name)
	localPath := filepath.Join(COVER_BASE, name+".png")
	dropboxPath := filepath.Join(DROPBOX_COVER_BASE, name+".png")
	if share, err := dropboxUpload(localPath, dropboxPath, false); err != nil {
		return ITunesImage{}, err
	} else {
		return ITunesImage{
			Href: share,
		}, nil
	}
}
