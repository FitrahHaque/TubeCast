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
	// station, err := user.CreateStation("test", "test-Station")
	station, err := rss.GetStation("test")
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	feedUrl, err := rss.GetChannelFeedUrl("@ThePrimeTimeagen")
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	// fmt.Printf("%v\n", feedUrl)
	if err = station.SyncChannel(feedUrl); err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	// station.Print()
}
