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
	station, err := rss.GetStation("test")
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	username := "@ThePrimeTimeagen"
	if err = station.SyncChannel(username); err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	// station.Print()
}
