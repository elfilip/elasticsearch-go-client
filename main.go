package main

import (
	"EsClient2/elastic"
	"EsClient2/store"
	"EsClient2/tab"
	"EsClient2/window"
	"encoding/json"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/container"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"sort"
	"strconv"
	"strings"
	"unicode"
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
	lConnections := createConnList(&config, tabs, &mainW)
	content := fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, lConnections, nil), lConnections, tabContainer)

	mainW.MainWindow.SetContent(content)
	mainW.MainWindow.ShowAndRun()
}

func createTab(connect *store.Connect, w *window.MainW) *widget.TabItem {
	progress := dialog.NewProgressInfinite("Connecting to ES", "Please wait...", w.MainWindow)
	progress.Show()
	elastic := elastic.NewElastic(connect)
	elastic.ConnectToES()
	indexTabs := widget.NewTabContainer()
	indexContent := fyne.NewContainerWithLayout(layout.NewMaxLayout(), indexTabs)
	indexList := createIndexList(indexTabs, elastic, connect, w)
	indexHeader := fyne.NewContainerWithLayout(layout.NewHBoxLayout(), widget.NewLabel(connect.Name), widget.NewLabel(connect.Url))

	content := fyne.NewContainerWithLayout(layout.NewBorderLayout(indexHeader, nil, indexList, nil),
		indexList, indexContent, indexHeader)

	tab := widget.NewTabItem(connect.Name, content)
	progress.Hide()
	//return widget.NewTabItem("Split", makeSplitTab())
	return tab
}

func createTabEditIndex(connect *store.Connect, w *window.MainW) *widget.TabItem {
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

func createIndexList(indexTabs *widget.TabContainer, elastic *elastic.Elastic, connect *store.Connect, w *window.MainW) *widget.ScrollContainer {
	indices := make([]fyne.CanvasObject, 0)
	indices = append(indices, widget.NewLabel("Indices"))
	for _, index := range elastic.Indices {
		indexInner := index
		idxButton := widget.NewButton(index.Name, func() {
			indexTab := createIndexContent(indexTabs, elastic, &indexInner, connect, w)
			indexTabs.Append(indexTab)
			indexTabs.SelectTab(indexTab)
			indexTabs.Refresh()
		})
		indices = append(indices, idxButton)
	}
	indicesLayout := widget.NewVScrollContainer(widget.NewVBox(indices...))
	return indicesLayout
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

func createIndexDataTable(data []map[string]interface{}, elastic *elastic.Elastic, index *elastic.Index,
connect *store.Connect, w *window.MainW) *widget.Table{
	var fields []string
	var prefFields []string
	for _,field := range connect.Preferred{
		if strings.Index(index.Name, field.IndexPrefix) == 0 {
			for _,fieldInner := range field.Fields{
				prefFields = append(prefFields, fieldInner)
			}
		}
	}
	if len(data) == 0 {
		return widget.NewTable(func() (int, int) {
			return 0,0
		}, func() fyne.CanvasObject {
			return widget.NewLabel("")
		}, func(id widget.TableCellID, object fyne.CanvasObject) {
			
		})
	}
	for key, _ :=range data[0]{
		fields = append(fields, key)
	}

	sort.Strings(fields)

	for _, val := range fields{
		prefFields = append(prefFields, val)
	}

	fields = prefFields

	table := widget.NewTable(func() (int, int) {
		return len(data)+1, len(fields)
	}, func() fyne.CanvasObject {
		return widget.NewLabel("randomaaaaaaa")
	}, func(id widget.TableCellID, object fyne.CanvasObject) {
		label := object.(*widget.Label)
		if id.Row == 0 {
			label.TextStyle = fyne.TextStyle{Bold: true}
		}else{
			label.TextStyle = fyne.TextStyle{}
		}
		var text string
		if id.Row != 0 {
			text = getStringFromESFirstLevelField(fields[id.Col], data[id.Row-1],false)
		} else{
			text = fields[id.Col]
		}
		if len(text) == 0 {
			label.SetText("")
		}else {
			cutText := Substr(text, 10)
			label.SetText(removeAccents(cutText))
		}
	})

	table.OnSelected = func(id widget.TableCellID) {
		if id.Row == 0 {
			return
		}
		fieldContent := getStringFromESFirstLevelField(fields[id.Col], data[id.Row-1],true)
		fieldEdit := widget.NewMultiLineEntry()
		fieldEdit.SetText(fieldContent)
		scroll :=container.NewScroll(fieldEdit)
		scroll.SetMinSize(fyne.NewSize(400,400))
		dialog.ShowCustomConfirm(fields[id.Col], "Save", "Cancel", scroll, func(b bool) {
			if !b {
				return
			}
		}, w.MainWindow)
	}

	return table
}

func getStringFromESFirstLevelField(field string, data interface{}, format bool) string{
	return 	getStringFromEsPath(strings.Split(field,"."), data, 0, format)
}

func getStringFromEsPath(path []string, data interface{}, index int, format bool) string {
	if index == len(path) {
		return convertAnyToString(data, format)
	}
	if data == nil {
		return ""
	}
	switch data.(type) {
	case map[string]interface{}:
		return getStringFromEsPath(path, data.(map[string]interface{})[path[index]], index+1, format)
	case []interface{}:
		return getStringFromEsPath(path, data.([]interface{})[0], index, format)
	default:
		return ""
	}
}

func createIndexContent(indexTabs *widget.TabContainer, elastic *elastic.Elastic, index *elastic.Index,
	connect *store.Connect, w *window.MainW) *widget.TabItem {

	indexTabs.Show()
	docMap :=elastic.LoadNFirstDocs(10, index.Name, "*")

	table := createIndexDataTable(docMap, elastic, index, connect, w)
	searchLabel := widget.NewLabel("Search: ")
	searchInput :=tab.NewSearchField()
	searchInput.SetText("*")

	searchLayout := fyne.NewContainerWithLayout(layout.NewHBoxLayout(), searchLabel, searchInput, layout.NewSpacer())
	vboxLayout := fyne.NewContainerWithLayout(layout.NewBorderLayout(searchLayout, nil, nil, nil), searchLayout, table)

	indexTab := widget.NewTabItem(index.Name, fyne.NewContainerWithLayout(layout.NewMaxLayout(),
		vboxLayout))
	closeTabButton := widget.NewButton("Close", func() {
		if len(indexTabs.Items) == 1 {
			indexTabs.Hide()
			return
		}
		if indexTabs.CurrentTabIndex() == len(indexTabs.Items) -1{
			indexTabs.SelectTabIndex(indexTabs.CurrentTabIndex()-1)
		}
		indexTabs.Remove(indexTab)
	})
	searchLayout.Add(closeTabButton)
	searchInput.EnterEvent = func() {
		vboxLayout.Objects[1] = createIndexDataTable(elastic.LoadNFirstDocs(20, index.Name, searchInput.Text), elastic, index, connect, w)
		vboxLayout.Refresh()
	}

	return indexTab
}

func convertAnyToString(field interface{}, format bool) string{
	var text string
	switch field.(type) {
	case map[string]interface{}:
		if format {
			val,_ := json.MarshalIndent(field.(map[string]interface{}), "", "  ")
			text = string(val)
		}else{
			val,_ := json.Marshal(field.(map[string]interface{}))
			text = string(val)
		}
	case []interface{}:
		if format {
			val,_:=json.MarshalIndent(field.([]interface{}),"", "  ")
			text = string(val)
		}else{
			val,_:=json.Marshal(field.([]interface{}))
			text = string(val)
		}
	case string:
		text = fmt.Sprintf("%s", field)
	case float64:
		text = fmt.Sprintf("%f", field)
		num := field.(float64)
		if  float64(int64(num)) == num{
			text = fmt.Sprintf("%d", int64(num))
		}
	case float32:
		text = fmt.Sprintf("%f", field)
		num := field.(float32)
		if  float32(int64(num)) == num{
			text = fmt.Sprintf("%d", int64(num))
		}
	default:
		text = fmt.Sprintf("%#v", field)
	}
	return text
}

func removeAccents(str string) string{
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
	noAccents, _, _ := transform.String(t, str)
	return noAccents
}

func Substr(str string, count int) string {
	var sb strings.Builder
	for i:=0; i < len(str) && i< count; i++ {
		sb.WriteByte(str[i])
	}
	return sb.String()
}

func createConnList(config *store.Config, tabs *widget.TabContainer, w *window.MainW) *fyne.Container {
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
					w.UpdateConnectionList(createConnList(config, tabs, w))
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
			w.UpdateConnectionList(createConnList(config, tabs, w))
		}), fyne.NewMenuItem("Edit", func() {
			tabs.Append(createTabEditIndex(fixedItem, w))
			tabs.SelectTabIndex(len(tabs.Items) -1)
			tabs.Refresh()
		}))
		elements = append(elements, newContextMenuButton(fixedItem.Name, contextMenu))
	}

	lConnections := fyne.NewContainerWithLayout(layout.NewVBoxLayout(), elements...)
	return lConnections
}

func createContent() (*fyne.Container, *widget.TabContainer) {
	tabs := widget.NewTabContainer()
	return fyne.NewContainerWithLayout(layout.NewMaxLayout(), tabs), tabs

}

func exit(app fyne.App) {
	app.Quit()
}

func showNextWindow(app fyne.App) {
	code := widget.NewTextGrid()
	code.ShowLineNumbers = true
	code.ShowWhitespace = false

	w := app.NewWindow("Next Window")
	w.SetContent(widget.NewVBox(widget.NewLabel("aaa"),
		widget.NewLabel("Look here!!!"),
		widget.NewButton("close!", func() {
			w.Hide()
		}),
		widget.NewTextGrid()))

	w.Resize(fyne.NewSize(100, 100))
	w.Show()
}

func showConnectionsWindows(app fyne.App, config store.Config) {
	w := app.NewWindow("Connections")
	w.SetFixedSize(true)
	elements := make([]fyne.CanvasObject, 0)
	elements = append(elements,
		widget.NewLabel(""),
		widget.NewLabel("Name"),
		widget.NewLabel("URL"))
	for index, item := range config.Connections {
		elements = append(elements,
			widget.NewLabel(strconv.Itoa(index)),
			widget.NewLabel(item.Name),
			widget.NewLabel(item.Url))
	}
	grid := fyne.NewContainerWithLayout(layout.NewGridLayout(3), elements...)
	w.SetContent(grid)
	w.Show()
}

type contextMenuButton struct {
	widget.Button
	menu *fyne.Menu
}

func (b *contextMenuButton) Tapped(e *fyne.PointEvent) {
	widget.ShowPopUpMenuAtPosition(b.menu, fyne.CurrentApp().Driver().CanvasForObject(b), e.AbsolutePosition)
}

func newContextMenuButton(label string, menu *fyne.Menu) *contextMenuButton {
	b := &contextMenuButton{menu: menu}
	b.Text = label

	b.ExtendBaseWidget(b)
	return b
}

func isMn(r rune) bool {
	return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
}
