package tab

import (
	"EsClient2/elastic"
	"EsClient2/service"
	"EsClient2/store"
	"EsClient2/util"
	"EsClient2/window"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/container"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
	"sort"
	"strings"
)

func createIndexContent(indexTabs *widget.TabContainer, elastic *elastic.Elastic, index *elastic.Index,
	alias *elastic.Alias, w *window.MainW) *widget.TabItem {

	indexTabs.Show()
	docMap := elastic.LoadNFirstDocs(10, index.Name, "*")

	table := createIndexDataTable(service.NewEsData(elastic, index, docMap, "*"), w)
	searchLabel := widget.NewLabel("Search: ")
	searchInput := NewSearchField()
	searchInput.SetText("*")
	searchLayout := fyne.NewContainerWithLayout(layout.NewHBoxLayout(), searchLabel, searchInput, layout.NewSpacer())
	headerLayout := fyne.NewContainerWithLayout(layout.NewVBoxLayout())
	vboxLayout := fyne.NewContainerWithLayout(layout.NewBorderLayout(headerLayout, nil, nil, nil), headerLayout, table)

	var tabName string
	if alias == nil {
		tabName = index.Name
	}else{
		indexName := widget.NewEntry()
		indexName.SetText(alias.IndexInner.Name)
		headerLayout.Add(fyne.NewContainerWithLayout(layout.NewHBoxLayout(),
			widget.NewLabel("Index name: "), indexName))
		tabName = alias.Alias
	}
	headerLayout.Add(searchLayout)
	indexTab := widget.NewTabItem(tabName, fyne.NewContainerWithLayout(layout.NewMaxLayout(),
		vboxLayout))
	closeTabButton := widget.NewButton("Close", func() {
		if len(indexTabs.Items) == 1 {
			indexTabs.Hide()
			return
		}
		if indexTabs.CurrentTabIndex() == len(indexTabs.Items)-1 {
			indexTabs.SelectTabIndex(indexTabs.CurrentTabIndex() - 1)
		}
		indexTabs.Remove(indexTab)
	})
	searchLayout.Add(closeTabButton)
	searchInput.EnterEvent = func() {
		vboxLayout.Objects[1] = createIndexDataTable(
			service.NewEsData(elastic, index, elastic.LoadNFirstDocs(20, index.Name, searchInput.Text), searchInput.Text), w)
		vboxLayout.Refresh()
	}

	return indexTab
}

func createIndexDataTable(data *service.EsData, w *window.MainW) *widget.Table {
	var fields []string
	var prefFields []string
	prefFields = append(prefFields, "")
	for _, field := range data.Elastic.Connect.Preferred {
		if strings.Index(data.Index.Name, field.IndexPrefix) == 0 {
			for _, fieldInner := range field.Fields {
				prefFields = append(prefFields, fieldInner)
			}
		}
	}
	if len(data.Data) == 0 {
		return widget.NewTable(func() (int, int) {
			return 0, 0
		}, func() fyne.CanvasObject {
			return widget.NewLabel("")
		}, func(id widget.TableCellID, object fyne.CanvasObject) {

		})
	}
	for key, _ := range data.Data[0]["_source"].(map[string]interface{}) {
		fields = append(fields, key)
	}

	sort.Strings(fields)

	for _, val := range fields {
		prefFields = append(prefFields, val)
	}

	fields = prefFields

	table := widget.NewTable(func() (int, int) {
		return len(data.Data) + 1, len(fields)
	}, func() fyne.CanvasObject {
		return widget.NewLabel("randomaaaaaaa")
	}, func(id widget.TableCellID, object fyne.CanvasObject) {
		label := object.(*widget.Label)
		if id.Row == 0 {
			label.TextStyle = fyne.TextStyle{Bold: true}
		} else {
			label.TextStyle = fyne.TextStyle{}
		}
		var text string
		if id.Row != 0 {
			text = data.GetStringFromESFirstLevelField(fields[id.Col], id.Row-1, false)
		} else {
			text = fields[id.Col]
		}
		if len(text) == 0 {
			label.SetText("")
		} else {
			cutText := util.Substr(text, 10)
			label.SetText(util.RemoveAccents(cutText))
		}
	})

	table.OnSelected = func(id widget.TableCellID) {
		if id.Row == 0 {
			return
		}
		fieldContent := data.GetStringFromESFirstLevelField(fields[id.Col], id.Row-1, true)
		fieldEdit := widget.NewMultiLineEntry()
		fieldEdit.SetText(fieldContent)
		scroll := container.NewScroll(fieldEdit)
		scroll.SetMinSize(fyne.NewSize(750, 400))
		dialog.ShowCustomConfirm(fields[id.Col], "Save", "Cancel", scroll, func(confirm bool) {
			if confirm {
				docId := data.GetDocId(id.Row - 1)
				if id.Col == 0 {
					docContent := fieldEdit.Text
					data.Elastic.UpdateDoc(docId, docContent, data.Index)
					fmt.Println("Updated doc ", docId)
					data.RefreshData()
				}else{
					dataString := data.UpdatePath(id.Row-1, fields[id.Col], fieldEdit.Text)
					data.Elastic.UpdateDoc(docId, dataString, data.Index)
				}
			}
			table.Unselect(id)
		}, w.MainWindow)
	}

	return table
}

func createTab(connect *store.Connect, w *window.MainW) *widget.TabItem {
	progress := dialog.NewProgressInfinite("Connecting to ES", "Please wait...", w.MainWindow)
	progress.Show()
	elastic := elastic.NewElastic(connect)
	elastic.ConnectToES()
	indexTabs := widget.NewTabContainer()
	indexContent := fyne.NewContainerWithLayout(layout.NewMaxLayout(), indexTabs)
	indexAliasTab := widget.NewTabContainer()
	indexList := createIndexList(indexTabs, elastic, connect, w)
	aliasList := createAliasList(indexTabs, elastic, connect, w)
	indexAliasTab.Append(widget.NewTabItem("Indices", indexList))
	indexAliasTab.Append(widget.NewTabItem("Aliases", aliasList))
	indexAliasTab.SelectTabIndex(0)
	indexHeader := fyne.NewContainerWithLayout(layout.NewHBoxLayout(), widget.NewLabel(connect.Name), widget.NewLabel(connect.Url))

	content := fyne.NewContainerWithLayout(layout.NewBorderLayout(indexHeader, nil, indexAliasTab, nil),
		indexAliasTab, indexContent, indexHeader)

	tab := widget.NewTabItem(connect.Name, content)
	progress.Hide()
	//return widget.NewTabItem("Split", makeSplitTab())
	return tab
}

func createIndexList(indexTabs *widget.TabContainer, elastic *elastic.Elastic, connect *store.Connect, w *window.MainW) *widget.ScrollContainer {
	indices := make([]fyne.CanvasObject, 0)
	for _, index := range elastic.Indices {
		indexInner := index
		idxButton := widget.NewButton(index.Name, func() {
			indexTab := createIndexContent(indexTabs, elastic, &indexInner, nil, w)
			indexTabs.Append(indexTab)
			indexTabs.SelectTab(indexTab)
			indexTabs.Refresh()
		})
		idxButton.Alignment = widget.ButtonAlignLeading
		indices = append(indices, idxButton)
	}
	indicesLayout := widget.NewVScrollContainer(widget.NewVBox(indices...))
	return indicesLayout
}

func createAliasList(indexTabs *widget.TabContainer, elastic *elastic.Elastic, connect *store.Connect, w *window.MainW) *widget.ScrollContainer {
	indices := make([]fyne.CanvasObject, 0)
	for _, alias := range elastic.Aliases {
		aliasInner := alias
		idxButton := widget.NewButton(alias.Alias, func() {
			indexTab := createIndexContent(indexTabs, elastic, &aliasInner.IndexInner, &aliasInner, w)
			indexTabs.Append(indexTab)
			indexTabs.SelectTab(indexTab)
			indexTabs.Refresh()
		})
		idxButton.Alignment = widget.ButtonAlignLeading
		indices = append(indices, idxButton)
	}
	indicesLayout := widget.NewVScrollContainer(widget.NewVBox(indices...))
	return indicesLayout
}
