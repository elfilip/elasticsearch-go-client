package window

import (
	"EsClient2/store"
	"fyne.io/fyne"
	"fyne.io/fyne/layout"
)

type MainW struct {
	MainWindow fyne.Window
	ConnectionList *fyne.Container
	Tabs *fyne.Container
	Config *store.Config
}

func (mainW *MainW) UpdateConnectionList(container *fyne.Container){
	mainW.ConnectionList = container
	content := fyne.NewContainerWithLayout(
		layout.NewBorderLayout(nil, nil, mainW.ConnectionList, nil), mainW.ConnectionList, mainW.Tabs)
	mainW.MainWindow.SetContent(content)
}