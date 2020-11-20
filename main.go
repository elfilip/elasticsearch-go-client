package main

import (
	"EsClient2/store"
	"EsClient2/tab"
	"EsClient2/window"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
)

func main() {
	a := app.New()

	mainW := window.MainW{}

	mainW.MainWindow = a.NewWindow("ES Client")
	config := store.NewStore()
	mainW.Config = &config
	mainW.MainWindow.Resize(fyne.NewSize(1000, 600))
	mainW.MainWindow.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("Quit", func() { exit(a) }),
			fyne.NewMenuItem("Settings", func() {})),
		fyne.NewMenu("Edit")))

	tabContainer, tabs := createContent()
	mainW.Tabs = tabContainer
	lConnections := tab.CreateConnList(&config, tabs, &mainW)
	content := fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, lConnections, nil), lConnections, tabContainer)

	mainW.MainWindow.SetContent(content)
	mainW.MainWindow.ShowAndRun()
}

func makeSplitTab() fyne.CanvasObject {
	left := widget.NewMultiLineEntry()
	left.Wrapping = fyne.TextWrapWord
	left.SetText("Long text is looooooooooooooong")
	right := widget.NewVSplitContainer(
		widget.NewLabel("Label"),
		widget.NewButton("Button", func() { fmt.Println("button tapped!") }),
	)
	return widget.NewHSplitContainer(widget.NewVScrollContainer(left), right)
}

func createContent() (*fyne.Container, *widget.TabContainer) {
	tabs := widget.NewTabContainer()
	return fyne.NewContainerWithLayout(layout.NewMaxLayout(), tabs), tabs

}

func exit(app fyne.App) {
	app.Quit()
}




