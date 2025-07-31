package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

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
	menu.Box.
		// SetBorder(true).
		SetTitle("MAIN MENU")
	menu.
		AddItem("shows", "Browse available shows", 'v', nil).
		AddItem("create a show", "Create a new show", 'c', nil).
		AddItem("subscribe", "Get latest videos from a YT channel easily", 's', nil).
		AddItem("sync", "Add latest episodes to all shows from your subscribed channels", 'a', nil).
		AddItem("add an episode", "Add an episode to a show", 'p', nil).
		// AddItem("remove an episode", "Remove an episode from a show", 'r', nil).
		AddItem("delete a show", "Remove a show", 'e', nil).
		AddItem("quit", "Exit the app", 'q', nil)

	// ==== LIST SHOWS =====
	shows := CreateShowsPage()
	// ===== PAGES ======
	pages := tview.NewPages()
	pages.
		AddPage("menu", menu, true, true).
		AddPage("shows", shows, true, false)

	// ==== Selection Callback Functions =====
	menu.SetSelectedFunc(func(_ int, mainText, _ string, _ rune) {
		switch mainText {
		case "shows":
			PopulateShows(shows)
			pages.SwitchToPage("shows")
		case "create a show":
			pages.AddAndSwitchToPage("create-show", CreateShowForm(application, pages), true)
		case "subscribe":
			pages.AddAndSwitchToPage("sync-channel", SubscribeForm(application, pages), true)
		case "sync":
			Sync(application, pages)
		case "add an episode":
			pages.AddAndSwitchToPage("add-video", AddEpisodeForm(application, pages), true)
		case "delete a show":
			pages.AddAndSwitchToPage("remove-show", RemoveShowForm(application, pages), true)
		case "quit":
			application.Stop()
		default:
			// no-op
		}
	})

	shows.SetSelectedFunc(func(_ int, mainText, _ string, _ rune) {
		switch mainText {
		case "← Back":
			pages.SwitchToPage("menu")
		default:
			// no-op
		}
	})
	// Start the application
	if err := application.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

// ShowSpinnerModal overlays a centered modal with an animated spinner.
// Call the returned stop() to remove it.
func ShowSpinnerModal(app *tview.Application, pages *tview.Pages, label string) (stop func()) {
	spinner := []rune{'|', '/', '-', '\\'}
	modal := tview.NewModal().SetText("⏳ " + label)
	modal.SetTitle(" Please wait ")

	done := make(chan struct{})
	go func() {
		app.QueueUpdateDraw(func() {
			pages.AddPage("loading", modal, true, true)
		})
	}()

	go func() {
		t := time.NewTicker(120 * time.Millisecond)
		defer t.Stop()
		for i := 0; ; {
			select {
			case <-done:
				return
			case <-t.C:
				i++
				ch := spinner[i%len(spinner)]
				app.QueueUpdateDraw(func() {
					modal.SetText(fmt.Sprintf("%s %c", label, ch))
				})
			}
		}
	}()

	return func() {
		close(done)
		app.QueueUpdateDraw(func() {
			pages.RemovePage("loading")
		})
	}
}

func Sync(app *tview.Application, pages *tview.Pages) {
	stop := ShowSpinnerModal(app, pages, "Syncing Shows...")
	go func() {
		err := rss.Sync()
		stop()
		app.QueueUpdateDraw(func() {
			var modal *tview.Modal
			if err != nil {
				modal = ShowModal(fmt.Sprintf("Syncing failed: %v", err), []string{"OK"}, func(_ int, _ string) {
					pages.RemovePage("modal")
				})
			} else {
				modal = ShowModal("Successfully synced", []string{"OK"}, func(_ int, _ string) {
					pages.RemovePage("modal")
				})
			}
			pages.AddPage("modal", modal, true, true)
		})
	}()
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
		// SetBorder(true).
		SetTitle(" Create a Show ")
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

func RemoveShowForm(app *tview.Application, pages *tview.Pages) tview.Primitive {
	titleIF := tview.NewInputField().
		SetLabel("Title:         ").
		SetFieldWidth(40)
	form := tview.NewForm()
	form.
		AddFormItem(titleIF).
		AddButton("remove", func() {
			title := titleIF.GetText()
			if title == "" {
				modal := ShowModal("Title is required to remove a show", []string{"OK"}, func(_ int, _ string) {
					pages.RemovePage("modal")
				})
				pages.AddPage("modal", modal, true, true)
			}
			err := rss.RemoveShow(title)
			var modal *tview.Modal
			if err == nil {
				modal = ShowModal(fmt.Sprintf("The show %s has been removed successfully", title), []string{"GO BACK"}, func(_ int, _ string) {
					pages.
						SwitchToPage("menu").
						RemovePage("modal")
				})
			} else {
				modal = ShowModal(fmt.Sprintf("%v", err), []string{"OK"}, func(_ int, _ string) {
					pages.RemovePage("modal")
				})
			}
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

func AddEpisodeForm(app *tview.Application, pages *tview.Pages) tview.Primitive {
	titleIF := tview.NewInputField().
		SetLabel("Title:         ").
		SetFieldWidth(40)
	descIF := tview.NewInputField().
		SetLabel("Description:   ").
		SetFieldWidth(80)
	urlIF := tview.NewInputField().
		SetLabel("Video Url:     ").
		SetFieldWidth(60)

	form := tview.NewForm()
	form.SetTitle(" Add an Episode ")
	form.
		AddFormItem(titleIF).
		AddFormItem(descIF).
		AddFormItem(urlIF).
		AddButton("Add", func() {
			title := titleIF.GetText()
			videoURL := urlIF.GetText()
			description := descIF.GetText()
			if title == "" || videoURL == "" {
				modal := ShowModal("Title and Video URL are required", []string{"OK"}, func(_ int, _ string) {
					pages.RemovePage("modal")
				})
				pages.AddPage("modal", modal, true, true)
			}
			stop := ShowSpinnerModal(app, pages, "Adding Episode...")
			go func() {
				result, err := rss.AddVideoToShow(title, description, videoURL)
				stop()
				app.QueueUpdateDraw(func() {

					var modal *tview.Modal
					if err == nil {
						modal = ShowModal(fmt.Sprintf("Video added successfully.\nShow link: %s", result), []string{"OK"}, func(_ int, _ string) {
							pages.
								SwitchToPage("menu").
								RemovePage("modal")
						})
					} else {
						modal = ShowModal(fmt.Sprintf("%v", err), []string{"OK"}, func(_ int, _ string) {
							pages.RemovePage("modal")
						})
					}
					pages.AddPage("modal", modal, true, true)
				})
			}()
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

func SubscribeForm(app *tview.Application, pages *tview.Pages) tview.Primitive {
	titleIF := tview.NewInputField().
		SetLabel("Title:         ").
		SetFieldWidth(40)
	descIF := tview.NewInputField().
		SetLabel("Description:   ").
		SetFieldWidth(80)
	channelIdIF := tview.NewInputField().
		SetLabel("Channel ID:    ").
		SetFieldWidth(20)
	form := tview.NewForm()
	form.SetTitle(" Add an Episode ")
	form.
		AddFormItem(titleIF).
		AddFormItem(descIF).
		AddFormItem(channelIdIF).
		AddButton("Add", func() {
			title := titleIF.GetText()
			channelId := channelIdIF.GetText()
			description := descIF.GetText()
			if title == "" || channelId == "" {
				modal := ShowModal("Title and Channel ID are required", []string{"OK"}, func(_ int, _ string) {
					pages.RemovePage("modal")
				})
				pages.AddPage("modal", modal, true, true)
			}
			stop := ShowSpinnerModal(app, pages, "Subcribing to the channel...")
			go func() {
				result, err := rss.SyncChannel(title, description, channelId)
				stop()
				app.QueueUpdateDraw(func() {

					var modal *tview.Modal
					if err == nil {
						modal = ShowModal(fmt.Sprintf("Channel Synced successfully.\nShow link: %s", result), []string{"OK"}, func(_ int, _ string) {
							pages.
								SwitchToPage("menu").
								RemovePage("modal")
						})
					} else {
						modal = ShowModal(fmt.Sprintf("%v", err), []string{"OK"}, func(_ int, _ string) {
							pages.RemovePage("modal")
						})
					}
					pages.AddPage("modal", modal, true, true)
				})
			}()
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

func CreateShowsPage() *tview.List {
	shows := tview.NewList()
	shows.
		// SetBorder(true).
		SetTitle("SHOWS")
	return PopulateShows(shows)
}

func PopulateShows(shows *tview.List) *tview.List {
	shows.Clear()
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
