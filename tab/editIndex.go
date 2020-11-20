package tab

import (
	"EsClient2/store"
	"EsClient2/window"
	"fyne.io/fyne"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

func CreateTabEditIndex(connect *store.Connect, w *window.MainW) *widget.TabItem {
	container := fyne.NewContainerWithLayout(layout.NewVBoxLayout())
	toolbar :=widget.NewToolbar()
	toolbar.Append(widget.NewToolbarAction(theme.FileIcon(), func() {
		indexPrefix := widget.NewEntry()
		indexPrefix.SetPlaceHolder("Enter index prefix")
		fields := widget.NewEntry()
		fields.SetPlaceHolder("Enter fields separated by comma")
		content := widget.NewForm(widget.NewFormItem("Index:", indexPrefix),
			widget.NewFormItem("fields:", fields))
		dialog.ShowCustomConfirm("New extra fields ...", "Create", "Cancel", content, func(b bool) {
			if b {
				connect.AddPreferedField(indexPrefix.Text, fields.Text)
				w.Config.Save()
				container.Add(createPrefFieldLine(indexPrefix.Text, fields.Text, len(connect.Preferred)+1,connect, container))
			}
			container.Refresh()
		}, w.MainWindow)
	}))
	nameLabel:=widget.NewLabel("Name: ")
	nameInput:=widget.NewEntry()
	nameInput.SetText(connect.Name)
	urlLabel:=widget.NewLabel("Url: ")
	urlInput :=widget.NewEntry()
	urlInput.SetText(connect.Url)
	container.Add(toolbar)
	container.Add(fyne.NewContainerWithLayout(layout.NewHBoxLayout(), nameLabel, nameInput))
	container.Add(fyne.NewContainerWithLayout(layout.NewHBoxLayout(), urlLabel, urlInput))
	for index, val := range connect.Preferred{
		container.Add(createPrefFieldLine(val.IndexPrefix, val.GetFieldsAsText(), index, connect, container))
	}

	tab := widget.NewTabItem("Edit " + connect.Name, container)
	return tab
}

func createPrefFieldLine(indexPrefix string, fields string, index int, connect *store.Connect, container *fyne.Container) *fyne.Container {
	prefIndex :=widget.NewEntry()
	prefIndex.SetText(indexPrefix)
	prefFields :=widget.NewEntry()
	prefFields.SetText(fields)
	prefHBox := fyne.NewContainerWithLayout(layout.NewHBoxLayout())
	prefDelete :=widget.NewButton("Delete", func() {
		connect.DeletePrefFieldByIndex(index)
		prefHBox.Hide()
		container.Refresh()
	})
	prefHBox.Add(prefIndex)
	prefHBox.Add(prefFields)
	prefHBox.Add(layout.NewSpacer())
	prefHBox.Add(prefDelete)
	return prefHBox
}
