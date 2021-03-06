package main

// Interfaces
// ----------
// Widget
//
// Structs
// -------
// LayoutItem
// Layout
//
// Consts
// ------
// WidgetEventNone
//
// LayoutItem
// ----------
// NewLayoutItem(widget Widget, visible bool) *LayoutItem
//
// Layout
// ------
// NewLayout() *Layout
// Rect() Rect
// AddItem(item *LayoutItem)
// SetFocusItem(item *LayoutItem)
// Draw()
// HandleEvent(e *tb.Event) (Widget, WidgetEventID)
//

import (
	tb "github.com/nsf/termbox-go"
)

type Widget interface {
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

func NewLayoutItem(widget Widget, visible bool) *LayoutItem {
	return &LayoutItem{
		Widget:  widget,
		Visible: visible,
	}
}

func NewLayout() *Layout {
	return &Layout{}
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
