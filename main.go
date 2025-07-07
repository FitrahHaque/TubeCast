package main

import (
	"fmt"

	"github.com/FitrahHaque/TubeCast/tubecast/rss"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	var user rss.User
	link, err := user.UploadToDropbox("./tubecast/audio/vDWaKVmqznQ.mp3", "/PodcastAudio/vDWaKVmqznQ.mp3")
	if err != nil {
		fmt.Printf("%v\n", err)
	} else {
		fmt.Printf("shareable link:%v\n", link)
	}
}
