package rss

import (
	"encoding/json"
	"os"
)

// Atomatically save Station Meta data locally
func saveStations(path string, stations []metaStation) error {
	tmp := path + ".tmp"

	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	defer f.Close()
	defer os.Remove(tmp)
	enc := json.NewEncoder(f)
	enc.SetIndent("", " ")
	if err := enc.Encode(stations); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Loads Station Meta data from the local
func loadStations(path string) ([]metaStation, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return []metaStation{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var stations []metaStation
	dec := json.NewDecoder(f)
	if err := dec.Decode(&stations); err != nil {
		return nil, err
	}
	return stations, nil
}
