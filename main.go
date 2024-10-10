package main

import (
	"io"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

type config struct {
	EditWidget    *widget.Entry
	PreviewWidget *widget.RichText
	CurrentFile   fyne.URI
	SaveMenuItem  *fyne.MenuItem
}

var cfg config

func main() {
	/*
		- create a fyne app
		- create a window for the app
		- get the user interface
	*/

	a := app.New()

	a.Settings().SetTheme(&myTheme{})
	win := a.NewWindow("Markdown Editor")

	edit, preview := cfg.makeUI()

	cfg.createMenuItems(win)

	// set content of the window
	win.SetContent(container.NewHSplit(edit, preview)) // split the window into two parts (edit | preview)

	win.Resize((fyne.Size{Width: 800, Height: 500}))
	win.CenterOnScreen() // center the window on the screen
	win.ShowAndRun()     // show the window and run the app
}

func (app *config) makeUI() (*widget.Entry, *widget.RichText) {
	edit := widget.NewMultiLineEntry()
	preview := widget.NewRichTextFromMarkdown("")

	cfg.EditWidget = edit
	cfg.PreviewWidget = preview

	edit.OnChanged = preview.ParseMarkdown

	return edit, preview

}

// createMenuItems creates the menu items for the app
func (app *config) createMenuItems(win fyne.Window) {
	openMenuItem := fyne.NewMenuItem("Open", app.openFunc(win))

	saveMenuItem := fyne.NewMenuItem("Save", app.saveFunc(win))
	app.SaveMenuItem = saveMenuItem
	app.SaveMenuItem.Disabled = true
	saveAsMenuItem := fyne.NewMenuItem("Save As", app.saveAsFunc(win))

	fileMenu := fyne.NewMenu("File", openMenuItem, saveMenuItem, saveAsMenuItem)

	menu := fyne.NewMainMenu(fileMenu)

	win.SetMainMenu(menu)
}

var filter = storage.NewExtensionFileFilter([]string{".md", ".MD"})

func (app *config) saveFunc(win fyne.Window) func() {
	return func() {
		if app.CurrentFile != nil {
			write, err := storage.Writer(app.CurrentFile)
			if err != nil {
				dialog.ShowError(err, win)
				return
			}

			write.Write([]byte(app.EditWidget.Text))
			defer write.Close()
		}
	}
}



// open file dialog
func (app *config) openFunc(win fyne.Window) func() {
	return func() {
		openDialog := dialog.NewFileOpen(func(read fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, win)
			}

			if read == nil {
				// user cancelled
				return
			}

			defer read.Close()
			// read file

			data, err := io.ReadAll(read)
			if err != nil {
				dialog.ShowError(err, win)
				return
			}

			app.EditWidget.SetText(string(data))

			app.CurrentFile = read.URI()

			win.SetTitle(win.Title() + " - " + read.URI().Name())
			app.SaveMenuItem.Disabled = false

		}, win)
		openDialog.SetFilter(filter)
		openDialog.Show()
	}
}

// ** try using the OS native file dialog to save the file
// this func uses fyne's dialog package to create a file save dialog
func (app *config) saveAsFunc(win fyne.Window) func() {
	return func() {
		saveDialog := dialog.NewFileSave(func(write fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, win)
				return
			}

			if write == nil {
				// user cancelled
				return
			}

			if !strings.HasSuffix(strings.ToLower(write.URI().String()), ".md") {
				dialog.ShowInformation("ERROR", "Please name your file with a .md extension", win)
				return
			}
			// save file
			write.Write([]byte(app.EditWidget.Text))
			app.CurrentFile = write.URI()

			defer write.Close()

			win.SetTitle(win.Title() + " - " + write.URI().Name())
			app.SaveMenuItem.Disabled = false

		}, win)

		if app.CurrentFile == nil || app.CurrentFile.Name() == "" {
			saveDialog.SetFileName("untitled.md")
		} else {
			saveDialog.SetFileName(app.CurrentFile.Name())
		}

		saveDialog.SetFilter(filter)
		saveDialog.Show()
	}
}
