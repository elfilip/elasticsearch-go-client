package tab

import (
	"EsClient2/store"
	"EsClient2/window"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

func CreateConnList(config *store.Config, tabs *widget.TabContainer, w *window.MainW) *fyne.Container {
	elements := make([]fyne.CanvasObject, 0)

	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.DocumentCreateIcon(), func() {
			name := widget.NewEntry()
			name.SetPlaceHolder("Type name for connection")
			url := widget.NewEntry()
			url.SetPlaceHolder("Type connection URL")
			content := widget.NewForm(widget.NewFormItem("Username:", name),
				widget.NewFormItem("URL:", url))
			dialog.ShowCustomConfirm("New Connection ...", "Create", "Cancel", content, func(b bool) {
				if b {
					newConnect := store.Connect{
						Name: name.Text,
						Url:  url.Text,
					}
					config.AddConnection(&newConnect)
					config.Save()
					w.UpdateConnectionList(CreateConnList(config, tabs, w))
				} else {
					fmt.Println("New connection cancelled")
				}
			}, w.MainWindow)
		}))
	elements = append(elements, toolbar)
	label := widget.NewLabelWithStyle("Connections        ", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	elements = append(elements, label)

	for index, _ := range config.Connections {
		fixedItem := &config.Connections[index]
		fixedIndex := index

		contextMenu := fyne.NewMenu(fixedItem.Name, fyne.NewMenuItem("Connect", func() {
			tabs.Append(createTab(fixedItem, w))
			tabs.SelectTabIndex(len(tabs.Items) - 1)
			fmt.Println("Tab")
			tabs.Refresh()
		}), fyne.NewMenuItem("Delete", func() {
			config.RemoveConnection(fixedIndex)
			w.UpdateConnectionList(CreateConnList(config, tabs, w))
		}), fyne.NewMenuItem("Edit", func() {
			tabs.Append(CreateTabEditIndex(fixedItem, w))
			tabs.SelectTabIndex(len(tabs.Items) -1)
			tabs.Refresh()
		}))
		elements = append(elements, NewContextMenuButton(fixedItem.Name, contextMenu))
	}

	lConnections := fyne.NewContainerWithLayout(layout.NewVBoxLayout(), elements...)
	return lConnections
}

