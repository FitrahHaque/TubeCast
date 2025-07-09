package main

import (
	"fmt"

	"github.com/FitrahHaque/TubeCast/tubecast/rss"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	rss.Init()
	// var user rss.User
	// station, err := user.CreateStation("test", "testing...")
	station, err := rss.GetStation("test-1")
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	username := "@ThePrimeTimeagen"
	if share, err := station.SyncChannel(username); err != nil {
		fmt.Printf("%v\n", err)
	} else {
		fmt.Printf("RSSFeed URL:\n%v\n", share)
	}
	// rss.UploadRSS("test-1")
	// station.Print()
}
