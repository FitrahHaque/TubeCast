package rss

import (
	"encoding/json"
	"encoding/xml"
	"os"
	"path/filepath"
)

// Atomatically save Station Meta data locally
func saveMetaStationToLocal(path string, station MetaStation) error {
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
	if err := enc.Encode(station); err != nil {
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
func saveXMLToLocal(path string, station Station) error {
	tmp := path + ".tmp"

	f, err := os.Create(tmp)
	if err != nil {
		return err
	}

	defer f.Close()
	defer os.Remove(tmp)

	enc := xml.NewEncoder(f)
	f.WriteString(xml.Header)
	if err := enc.Encode(station); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Loads Station data from the local
func loadXMLfromLocal(path string) (Station, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return Station{}, nil
	}
	if err != nil {
		return Station{}, err
	}

	defer f.Close()

	var station Station
	dec := xml.NewDecoder(f)
	if err := dec.Decode(&station); err != nil {
		return Station{}, err
	}
	return station, nil
}

// Save Station to cloud
// func UploadStation(station Station) error {

// }

// // Fetch Station from cloud
// func FetchStation(url string) (Station, error) {

// }
