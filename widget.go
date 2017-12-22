package main

import (
	tb "github.com/nsf/termbox-go"
)

type Widget interface {
	Rect() Rect
	Draw()
	HandleEvent(e *tb.Event) (Widget, WidgetEventID)
}

type WidgetEventID uint

const (
	WidgetEventNone WidgetEventID = iota
)

type LayoutItem struct {
	Widget
	Visible bool
}

type Layout struct {
	Items     []*LayoutItem
	FocusItem *LayoutItem
}

func NewLayout() *Layout {
	return &Layout{}
}

func NewLayoutItem(widget Widget, visible bool) *LayoutItem {
	return &LayoutItem{
		Widget:  widget,
		Visible: visible,
	}
}

func (layout *Layout) Rect() Rect {
	return NewRect(0, 0, 0, 0)
}

func (layout *Layout) AddItem(item *LayoutItem) {
	layout.Items = append(layout.Items, item)
}
func (layout *Layout) SetFocusItem(item *LayoutItem) {
	layout.FocusItem = item
}

func (layout *Layout) Draw() {
	for _, item := range layout.Items {
		if item.Visible {
			item.Widget.Draw()
		}
	}
}

func (layout *Layout) HandleEvent(e *tb.Event) (Widget, WidgetEventID) {
	if layout.FocusItem == nil {
		return layout, WidgetEventNone
	}
	return layout.FocusItem.HandleEvent(e)
}
