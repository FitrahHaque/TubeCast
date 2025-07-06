package rss

import (
	"encoding/json"
	"os"
)

// Atomatically save Station Meta data locally
func saveMetaStationToLocal(path string, station MetaStation) error {
	tmp := path + ".tmp"

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
