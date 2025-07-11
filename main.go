package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/FitrahHaque/TubeCast/tubecast/rss"
	"github.com/joho/godotenv"
)

var Commands = [...]string{"sync-channel", "create-show", "sync", "add-video", "help"}

func main() {
	godotenv.Load()
	rss.Init()
	application := os.Args[0]
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	syncChannelCmd := flag.Bool(Commands[0], false, "Sync a specific channel")
	createShowCmd := flag.Bool(Commands[1], false, "Create a new show")
	syncCmd := flag.Bool(Commands[2], false, "Sync all shows")
	addVideoCmd := flag.Bool(Commands[3], false, "Add video to show")
	helpCmd := flag.Bool(Commands[4], false, "Help")

	if len(os.Args) == 1 {
		fmt.Println("Please provide commands")
		os.Exit(1)
	}
	commandArgs := findIntersection(
		[]string{
			"-sync-channel",
			"-create-show",
			"-sync",
			"-add-video",
		},
		os.Args[1:2],
	)

	flag.CommandLine.Parse(commandArgs)

	commandsSelected := countTrue([]bool{*syncChannelCmd, *createShowCmd, *syncCmd, *addVideoCmd})

	if commandsSelected > 1 {
		fmt.Println("Specify a single command")
		os.Exit(1)
	} else if commandsSelected == 0 {
		commandArgs = findIntersection(
			[]string{
				"-help",
			},
			os.Args[1:],
		)

		flag.CommandLine.Parse(commandArgs)
		if *helpCmd {
			fmt.Fprintf(os.Stderr, "Usage of %s:\n", application)
			fmt.Fprintf(os.Stderr, "Valid commands include:\n\t%s\n", strings.Join(Commands[:], ", "))
			fmt.Fprintf(os.Stderr, "Flag:\n")
			flag.PrintDefaults()
			return
		}

		fmt.Println("No command is selected")
		os.Exit(1)
	}

	checkForSyncChannel(application, syncChannelCmd, 1)
	checkForCreateShow(application, createShowCmd, 1)
	checkForSync(syncCmd)
	checkForAddVideo(application, addVideoCmd, 1)
}

func countTrue(commands []bool) int {
	count := 0
	for _, c := range commands {
		if c == true {
			count++
		}
	}
	return count
}

func checkForSyncChannel(application string, syncChannelCmd *bool, cmdIdx int) {
	if *syncChannelCmd {
		syncChannelFS := flag.NewFlagSet("sync-channel", flag.ExitOnError)
		syncChannelFS.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s -sync-channel [OPTIONS]\n", application)
			fmt.Fprintf(os.Stderr, "Valid options include:\n\t%s\n", strings.Join([]string{"title, description, channel-id, help"}, ", "))
			fmt.Fprintf(os.Stderr, "Flag:\n")
			syncChannelFS.PrintDefaults()
		}

		title := syncChannelFS.String("title", "", "Title of the show")
		description := syncChannelFS.String("description", "", "Description of the show")
		channelID := syncChannelFS.String("channel-id", "", "Channel ID to sync")
		help := syncChannelFS.Bool("help", false, "Help")

		commandArgs := findIntersection(
			[]string{
				"--title",
				"--description",
				"--channel-id",
				"--help",
			},
			os.Args[cmdIdx+1:],
		)

		syncChannelFS.Parse(commandArgs)

		if *help {
			syncChannelFS.Usage()
			return
		}

		if *title == "" || *channelID == "" {
			fmt.Println("Title and Channel ID are required")
			syncChannelFS.Usage()
			os.Exit(1)
		}

		result, err := rss.SyncChannel(*title, *description, *channelID)
		if err != nil {
			fmt.Printf("Error syncing channel: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Channel synced successfully.\nPaste this RSS Feed link to your Podcast App: %s\n", result)
	}
}

func checkForCreateShow(application string, createShowCmd *bool, cmdIdx int) {
	if *createShowCmd {
		createShowFS := flag.NewFlagSet("create-show", flag.ExitOnError)
		createShowFS.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s create-show [OPTIONS]\n", application)
			fmt.Fprintf(os.Stderr, "Valid options include:\n\t%s\n", strings.Join([]string{"title, description, help"}, ", "))
			fmt.Fprintf(os.Stderr, "Flag:\n")
			createShowFS.PrintDefaults()
		}

		title := createShowFS.String("title", "", "Title of the show")
		description := createShowFS.String("description", "", "Description of the show")
		help := createShowFS.Bool("help", false, "Help")

		commandArgs := findIntersection(
			[]string{
				"--title",
				"--description",
				"--help",
			},
			os.Args[cmdIdx+1:],
		)

		createShowFS.Parse(commandArgs)

		if *help {
			createShowFS.Usage()
			return
		}

		if *title == "" || *description == "" {
			fmt.Println("Title and Description are required to create a new show")
			createShowFS.Usage()
			os.Exit(1)
		}

		result := rss.CreateShow(*title, *description)
		fmt.Printf("Show created successfully.\nPaste this RSS Feed link to your Podcast App: %s\n", result)
	}
}

func checkForSync(syncCmd *bool) {
	if *syncCmd {
		err := rss.Sync()
		if err != nil {
			fmt.Printf("Error syncing: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("All shows synced successfully")
	}
}

func checkForAddVideo(application string, addVideoCmd *bool, cmdIdx int) {
	if *addVideoCmd {
		addVideoFS := flag.NewFlagSet("add-video", flag.ExitOnError)
		addVideoFS.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s add-video [OPTIONS]\n", application)
			fmt.Fprintf(os.Stderr, "Valid options include:\n\t%s\n", strings.Join([]string{"title, description, video-url, help"}, ", "))
			fmt.Fprintf(os.Stderr, "Flag:\n")
			addVideoFS.PrintDefaults()
		}

		title := addVideoFS.String("title", "", "Title of the show")
		description := addVideoFS.String("description", "", "Description of the show")
		videoURL := addVideoFS.String("video-url", "", "Video URL to add")
		help := addVideoFS.Bool("help", false, "Help")

		commandArgs := findIntersection(
			[]string{
				"--title",
				"--description",
				"--video-url",
				"--help",
			},
			os.Args[cmdIdx+1:],
		)

		addVideoFS.Parse(commandArgs)

		if *help {
			addVideoFS.Usage()
			return
		}

		if *title == "" || *videoURL == "" {
			fmt.Println("Title and Video URL are required")
			addVideoFS.Usage()
			os.Exit(1)
		}

		result, err := rss.AddVideoToShow(*title, *description, *videoURL)
		if err != nil {
			fmt.Printf("Error adding video: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Video added successfully.\nPaste this RSS Feed link to your Podcast App: %s\n", result)
	}
}

func findIntersection(commandList, argList []string) []string {
	set := rss.NewSet[string]()
	for _, c := range commandList {
		set.Add(c)
	}
	var out []string
	for _, arg := range argList {
		cmd := arg
		if strings.Contains(cmd, "=") {
			cmd = strings.SplitN(cmd, "=", 2)[0]
		}

		if set.Has(cmd) {
			out = append(out, arg)
		}
	}
	return out
}
