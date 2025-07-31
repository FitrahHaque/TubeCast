package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/FitrahHaque/TubeCast/tubecast/rss"
	"github.com/atotto/clipboard"
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
		AddItem("add an episode", "Add an episode to a show", 'i', nil).
		AddItem("delete a show", "Remove a show", 'd', nil).
		AddItem("set env", "Set environment variables", 'e', nil).
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
			pages.AddAndSwitchToPage("remove-show", RemoveShow(application, pages), true)
		case "set env":
			pages.AddAndSwitchToPage("set-env", SetEnv(application, pages), true)
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
			ListEpisodes(application, pages, mainText)
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
		SetDoneFunc(cb).
		SetFocus(0)
	return modal
}

func CreateShowForm(app *tview.Application, pages *tview.Pages) tview.Primitive {
	titleIF := tview.NewInputField().
		SetLabel("Title:                  ").
		SetFieldWidth(30)
	descIF := tview.NewInputField().
		SetLabel("Description:            ").
		SetFieldWidth(80)
	coverImageIF := tview.NewInputField().
		SetLabel("Cover Image File Path:  ").
		SetFieldWidth(30)
	// titleIF.
	// 	SetDoneFunc(func(key tcell.Key) {
	// 		if key == tcell.KeyEnter {
	// 			app.SetFocus(descIF)
	// 		}
	// 	})

	form := tview.NewForm()
	form.
		// SetBorder(true).
		SetTitle(" Create a Show ")
	form.
		AddFormItem(titleIF).
		AddFormItem(descIF).
		AddFormItem(coverImageIF).
		AddButton("save", func() {
			title := titleIF.GetText()
			description := descIF.GetText()
			coverPath := coverImageIF.GetText()
			if title == "" || description == "" {
				modal := ShowModal("Title and Description are required to create a new show", []string{"OK"}, func(_ int, _ string) {
					pages.RemovePage("modal")
				})
				pages.AddPage("modal", modal, true, true)
			}
			stop := ShowSpinnerModal(app, pages, "Creating Show...")
			go func() {
				result, err := rss.CreateShow(title, description, coverPath)
				stop()
				app.QueueUpdateDraw(func() {
					var modal *tview.Modal
					if err != nil {
						modal = ShowModal(fmt.Sprintf("%v", err), []string{"OK"}, func(_ int, _ string) {
							pages.RemovePage("modal")
						})
					} else {
						modal = ShowModal(fmt.Sprintf("Show created successfully. Paste this link on your podcast app: %s", result), []string{"OK", "COPY SHOW LINK"}, func(_ int, label string) {
							switch label {
							case "COPY SHOW LINK":
								clipboard.WriteAll(result)
							}
							pages.
								SwitchToPage("menu").
								RemovePage("modal")
						})
					}
					pages.AddPage("modal", modal, true, true)
				})
			}()

		}).
		AddButton("cancel", func() {
			pages.SwitchToPage("menu")
		}).
		SetFocus(0)
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			pages.SwitchToPage("menu")
			return nil
		}
		return event
	})

	return form
}

func RemoveShow(app *tview.Application, pages *tview.Pages) tview.Primitive {
	shows := tview.NewList()
	shows.SetTitle("SHOWS")
	shows.AddItem("← Back", "Return To Main Menu", 'b', nil)
	for name := range rss.StationNames.Value {
		shows.AddItem(name, "Select to delete", 0, nil)
	}
	shows.SetSelectedFunc(func(_ int, mainText, _ string, _ rune) {
		switch mainText {
		case "← Back":
			pages.SwitchToPage("menu")
		default:
			stop := ShowSpinnerModal(app, pages, "Deleting Show...")
			go func() {
				err := rss.RemoveShow(mainText)
				stop()
				app.QueueUpdateDraw(func() {
					var modal *tview.Modal
					if err == nil {
						modal = ShowModal(fmt.Sprintf("The show %s has been removed successfully", mainText), []string{"GO BACK"}, func(_ int, _ string) {
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
		}
	})
	return shows
}
func AddEpisodeForm(app *tview.Application, pages *tview.Pages) tview.Primitive {
	titleIF := tview.NewInputField().
		SetLabel("Title:         ").
		SetFieldWidth(40)
	urlIF := tview.NewInputField().
		SetLabel("Video Url:     ").
		SetFieldWidth(60)

	form := tview.NewForm()
	form.SetTitle(" Add an Episode ")
	form.
		AddFormItem(titleIF).
		AddFormItem(urlIF).
		AddButton("Add", func() {
			title := titleIF.GetText()
			videoURL := urlIF.GetText()
			if title == "" || videoURL == "" {
				modal := ShowModal("Title and Video URL are required", []string{"OK"}, func(_ int, _ string) {
					pages.RemovePage("modal")
				})
				pages.AddPage("modal", modal, true, true)
			}
			stop := ShowSpinnerModal(app, pages, "Adding Episode...")
			go func() {
				result, err := rss.AddVideoToShow(title, videoURL)
				stop()
				app.QueueUpdateDraw(func() {
					var modal *tview.Modal
					if err == nil {
						modal = ShowModal(fmt.Sprintf("Video added successfully.\nShow link: %s", result), []string{"OK", "COPY SHOW LINK"}, func(_ int, label string) {
							switch label {
							case "COPY SHOW LINK":
								clipboard.WriteAll(result)
							}
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
		}).
		SetFocus(0)
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
	channelIdIF := tview.NewInputField().
		SetLabel("Channel ID:    ").
		SetFieldWidth(20)
	form := tview.NewForm()
	form.SetTitle(" Add an Episode ")
	form.
		AddFormItem(titleIF).
		AddFormItem(channelIdIF).
		AddButton("Add", func() {
			title := titleIF.GetText()
			channelId := channelIdIF.GetText()
			if title == "" || channelId == "" {
				modal := ShowModal("Title and Channel ID are required", []string{"Try Again"}, func(_ int, _ string) {
					pages.RemovePage("modal")
				})
				pages.AddPage("modal", modal, true, true)
			}
			stop := ShowSpinnerModal(app, pages, "Subcribing to the channel...")
			go func() {
				result, err := rss.SyncChannel(title, channelId)
				stop()
				app.QueueUpdateDraw(func() {

					var modal *tview.Modal
					if err == nil {
						modal = ShowModal(fmt.Sprintf("Channel Synced successfully.\nShow link: %s", result), []string{"OK", "COPY SHOW LINK"}, func(_ int, label string) {
							switch label {
							case "COPY SHOW LINK":
								clipboard.WriteAll(result)
							}
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
		}).
		SetFocus(0)
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
		shows.AddItem(name, "Browse episodes", 0, nil)
	}
	return shows
}

func ListEpisodes(app *tview.Application, pages *tview.Pages, showTitle string) {
	episodeInfos, err := rss.GetAllShowEpisodes(showTitle)
	if err != nil {
		modal := ShowModal(fmt.Sprint(err), []string{"OK"}, func(_ int, _ string) {
			pages.RemovePage("modal")
		})
		pages.AddPage("modal", modal, true, true)
		return
	}
	episodes := tview.NewList()
	episodes.SetTitle(fmt.Sprintf("Show %s Episodes", showTitle))
	episodes.AddItem("← Back", "Return to the Shows", 'b', nil)
	if len(episodeInfos) == 0 {
		episodes.AddItem(" No Episode Found ", "", 0, nil)
	} else {
		episodes.AddItem("▶ Copy link to clipboard", "Paste it on your podcast app", 'c', nil)
		episodes.AddItem("========= All Episodes =========", "", 0, nil)
		for _, info := range episodeInfos {
			episodes.AddItem(info.Title, info.Author, 0, nil)
		}
	}
	episodes.SetSelectedFunc(func(_ int, mainText, secondaryText string, _ rune) {
		switch mainText {
		case "← Back":
			pages.SwitchToPage("shows")
		case " No Episode Found ":
			// no-op
		case "========= All Episodes =========":
			// no-op
		case "▶ Copy link to clipboard":
			clipboard.WriteAll(rss.GetFeedUrl(showTitle))
		default:
			// episode
			RemoveEpisode(app, pages, showTitle, mainText, secondaryText)
		}
	})
	pages.AddAndSwitchToPage("episodes", episodes, true)
}

func RemoveEpisode(app *tview.Application, pages *tview.Pages, showTitle, episodeTitle, author string) {
	modal := ShowModal(fmt.Sprintf("Do you want to delete the episode titled \"%s\" from the show?", episodeTitle), []string{"NO", "YES"}, func(_ int, label string) {
		switch label {
		case "NO":
			pages.RemovePage("modal")
		case "YES":
			pages.RemovePage("modal")
			stop := ShowSpinnerModal(app, pages, "Deleting episode...")
			go func() {
				err := rss.RemoveVideoFromShow(showTitle, episodeTitle, author)
				stop()
				app.QueueUpdateDraw(func() {
					var modal *tview.Modal
					if err == nil {
						modal = ShowModal("Episode removed successfully", []string{"OK"}, func(_ int, _ string) {
							pages.RemovePage("modal")
							ListEpisodes(app, pages, showTitle)
						})
					} else {
						modal = ShowModal(fmt.Sprintf("%v", err), []string{"OK"}, func(_ int, _ string) {
							pages.RemovePage("modal")
						})
					}
					pages.AddPage("modal", modal, true, true)
				})
			}()
		}
	})
	pages.AddPage("modal", modal, true, true)
}

func SetEnv(app *tview.Application, pages *tview.Pages) tview.Primitive {
	usernameIF := tview.NewInputField().
		SetLabel("USERNAME:         ").
		SetFieldWidth(20).
		SetPlaceholder("A Unique Handle")
	homeIF := tview.NewInputField().
		SetLabel("HOME_DIR:         ").
		SetFieldWidth(20).
		SetPlaceholder("Your Home Directory")
	form := tview.NewForm()
	form.
		SetTitle(" Set Environment Variable ")
	form.
		AddFormItem(usernameIF).
		AddFormItem(homeIF).
		AddButton("save", func() {
			username := usernameIF.GetText()
			homeDir := homeIF.GetText()
			if username == "" || homeDir == "" || !strings.HasPrefix(homeDir, "/") {
				modal := ShowModal("username and home directory cannot be empty. Home directory needs to be from the root (absolute)", []string{"Try Again"}, func(_ int, _ string) {
					pages.RemovePage("modal")
				})
				pages.AddPage("modal", modal, true, true)
			} else {
				template, err := os.ReadFile("example.txt")
				if err != nil {
					modal := ShowModal(fmt.Sprint(err), []string{"OK"}, func(_ int, _ string) {
						pages.RemovePage("modal")
					})
					pages.AddPage("modal", modal, true, true)
				} else {
					lines := strings.Split(string(template), "\n")
					for i, line := range lines {
						if strings.HasPrefix(line, "USERNAME=") {
							lines[i] = fmt.Sprintf("USERNAME=\"%s\"", username)
						} else if strings.HasPrefix(line, "HOME_DIR=") {
							lines[i] = fmt.Sprintf("HOME_DIR=\"%s\"", homeDir)
						}
					}
					output := strings.Join(lines, "\n")
					err := os.WriteFile(".env", []byte(output), 0o644)
					if err != nil {
						modal := ShowModal(fmt.Sprint(err), []string{"OK"}, func(_ int, _ string) {
							pages.RemovePage("modal")
						})
						pages.AddPage("modal", modal, true, true)
					}
					modal := ShowModal("Successfully variables set", []string{"GO BACK"}, func(_ int, _ string) {
						pages.
							SwitchToPage("menu").
							RemovePage("modal")
					})
					pages.AddPage("modal", modal, true, true)
				}
			}
		}).
		AddButton("cancel", func() {
			pages.SwitchToPage("menu")
		}).
		SetFocus(0)
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			pages.SwitchToPage("menu")
			return nil
		}
		return event
	})
	return form

}
