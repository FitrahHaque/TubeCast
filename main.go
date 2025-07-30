package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/FitrahHaque/TubeCast/tubecast/rss"
	"github.com/gdamore/tcell/v2"
	"github.com/joho/godotenv"
	"github.com/rivo/tview"
)

var Commands = [...]string{"sync-channel", "create-show", "sync", "add-video", "remove-video", "remove-show", "help"}

func main() {
	godotenv.Load()
	rss.Init()
	application := tview.NewApplication().SetTitle("TubeCast")

	// ====== MENU =========
	menu := tview.NewList()
	menu.
		SetBorder(true).
		SetTitle("MAIN MENU")
	menu.
		AddItem("shows", "Browse available shows", 'v', nil).
		AddItem("create a show", "Create a new show", 'c', nil).
		AddItem("subscribe", "Get latest videos from a YT channel easily", 's', nil).
		AddItem("sync", "Add latest episodes to all shows from your subscribed channels", 'a', nil).
		AddItem("add an episode", "Add an episode to a show", 'p', nil).
		AddItem("remove an episode", "Remove an episode from a show", 'r', nil).
		AddItem("delete a show", "Remove a show", 'e', nil)

	// ===== List Shows =====
	shows := ListShows()

	// ===== PAGES ======
	pages := tview.NewPages()
	pages.
		AddPage("menu", menu, true, true).
		AddPage("create-show", CreateShowForm(application, pages), true, false).
		AddPage("shows", shows, true, false)

	if err := application.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
		log.Printf("Application exited with error: %v\n", err)
		os.Exit(1)
	}

	// ==== Selection Callback Functions =====
	menu.SetSelectedFunc(func(_ int, mainText, _ string, shortcut rune) {
		switch mainText {
		case "shows":
			pages.SwitchToPage("shows")
		case "create a show":
			pages.SwitchToPage("create-show")
		default:
			// no-op
		}
	})
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

func ShowModal(message string, buttonLabels []string, cb func(int, string)) *tview.Modal {
	modal := tview.NewModal().
		SetText(message).
		AddButtons(buttonLabels).
		SetDoneFunc(cb)
	return modal
}

func CreateShowForm(app *tview.Application, pages *tview.Pages) tview.Primitive {
	titleIF := tview.NewInputField().
		SetLabel("Title:         ").
		SetFieldWidth(40)
	descIF := tview.NewInputField().
		SetLabel("Description:   ").
		SetFieldWidth(80)
	titleIF.
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				app.SetFocus(descIF)
			}
		})

	form := tview.NewForm()
	form.
		SetBorder(true).
		SetTitle(" Create a Show ").
		SetTitleAlign(tview.AlignLeft)
	form.
		AddFormItem(titleIF).
		AddFormItem(descIF).
		AddButton("save", func() {
			title := titleIF.GetText()
			description := descIF.GetText()
			if title == "" || description == "" {
				modal := ShowModal("Title and Description are required to create a new show", []string{"OK"}, func(_ int, _ string) {
					pages.RemovePage("modal")
				})
				pages.AddPage("modal", modal, true, true)
			}
			result := rss.CreateShow(title, description)
			// fmt.Printf("Show created successfully.\nPaste this RSS Feed link to your Podcast App: %s\n", result)
			modal := ShowModal(fmt.Sprintf("Show created successfully.\nPaste this RSS Feed link to your Podcast App: %s\n", result), []string{"OK"}, func(_ int, _ string) {
				pages.
					SwitchToPage("menu").
					RemovePage("modal")
			})
			pages.AddPage("modal", modal, true, true)
		}).
		AddButton("cancel", func() {
			pages.SwitchToPage("menu")
		})
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			pages.SwitchToPage("menu")
			return nil
		}
		return event
	})

	return form
}

func ListShows() tview.Primitive {
	shows := tview.NewList()
	shows.
		SetBorder(true).
		SetTitle("SHOWS")
	shows.AddItem("← Back", "Return To Main Menu", 'b', nil)
	for name := range rss.StationNames.Value {
		shows.AddItem(name, fmt.Sprintf("Browse episodes from %s", name), 0, nil)
	}

	return shows
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
		fmt.Printf("sync command starts here....\n")
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

func checkForRemoveVideo(application string, removeVideoCmd *bool, cmdIdx int) {
	if *removeVideoCmd {
		removeVideoFS := flag.NewFlagSet("remove-video", flag.ExitOnError)
		removeVideoFS.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s -remove-video [OPTIONS]\n", application)
			fmt.Fprintf(os.Stderr, "Valid options include:\n\t%s\n", strings.Join([]string{"title, video-url, help"}, ", "))
			fmt.Fprintf(os.Stderr, "Flag:\n")
			removeVideoFS.PrintDefaults()
		}

		title := removeVideoFS.String("title", "", "Title of the show")
		videoURL := removeVideoFS.String("video-url", "", "Video URL to add")
		help := removeVideoFS.Bool("help", false, "Help")

		commandArgs := findIntersection(
			[]string{
				"--title",
				"--video-url",
				"--help",
			},
			os.Args[cmdIdx+1:],
		)

		removeVideoFS.Parse(commandArgs)

		if *help {
			removeVideoFS.Usage()
			return
		}

		if *title == "" || *videoURL == "" {
			fmt.Println("Title and Video URL are required")
			removeVideoFS.Usage()
			os.Exit(1)
		}

		err := rss.RemoveVideoFromShow(*title, *videoURL)
		if err != nil {
			fmt.Printf("Error adding video: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Video removed successfully.\n")
	}
}
func checkForRemoveShow(application string, removeShowCmd *bool, cmdIdx int) {
	if *removeShowCmd {
		removeShowFS := flag.NewFlagSet("remove-show", flag.ExitOnError)
		removeShowFS.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s -remove-show [OPTIONS]\n", application)
			fmt.Fprintf(os.Stderr, "Valid options include:\n\t%s\n", strings.Join([]string{"title, help"}, ", "))
			fmt.Fprintf(os.Stderr, "Flag:\n")
			removeShowFS.PrintDefaults()
		}

		title := removeShowFS.String("title", "", "Title of the show")
		help := removeShowFS.Bool("help", false, "Help")

		commandArgs := findIntersection(
			[]string{
				"--title",
				"--help",
			},
			os.Args[cmdIdx+1:],
		)

		removeShowFS.Parse(commandArgs)

		if *help {
			removeShowFS.Usage()
			return
		}

		if *title == "" {
			fmt.Println("Title is required")
			removeShowFS.Usage()
			os.Exit(1)
		}

		err := rss.RemoveShow(*title)
		if err != nil {
			fmt.Printf("Error adding video: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Show removed successfully.\n")
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
