package tab

import (
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/widget"
	"log"
)

func NewSearchField() *SearchField {
	search := widget.NewEntry()
	search.SetPlaceHolder(" Search")
	return &SearchField{search, nil}
}

type SearchField struct {
	*widget.Entry
	EnterEvent func()
}

func (s *SearchField) KeyUp(k *fyne.KeyEvent) {
	switch k.Name {
	case fyne.KeyReturn:
		log.Print("keyUP", fmt.Sprint("%h", k.Name))
		s.EnterEvent()
	}
	// s.Field.KeyUp(k)
	// widget.Refresh(s)
}

func (s *SearchField) CreateRenderer() fyne.WidgetRenderer {
	return widget.Renderer(s.Entry)
}
