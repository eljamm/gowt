package tui

import (
	"github.com/gdamore/tcell/v2"
)

type Action int

const (
	ActionNone Action = iota
	ActionDraw
	ActionQuit
	ActionSelect
)

type KeyEvent struct {
	Key  tcell.Key
	Rune rune
}

type RenderRequest struct {
	ModeIndicator string
	ModeStyle     tcell.Style
	Query         string
	Items         []WorktreeInfo
	Selected      int
	StatusPrompt  string
	NavDelta      int
	NavTop        bool
	NavBottom     bool
	NavTarget     int
	NavPage       int
}

type WorktreeInfo struct {
	Name    string
	AbsPath string
	Display string
	Commit  string
	Branch  string
	IsCwd   bool
}

type State interface {
	HandleKey(e KeyEvent) (State, Action, RenderRequest)
}

type InsertState struct {
	Query string
}

func (s InsertState) HandleKey(e KeyEvent) (State, Action, RenderRequest) {
	switch e.Key {
	case tcell.KeyUp, tcell.KeyCtrlP, tcell.KeyCtrlK:
		return s, ActionDraw, RenderRequest{Query: s.Query, ModeIndicator: " I ", NavDelta: -1}
	case tcell.KeyDown, tcell.KeyCtrlN, tcell.KeyCtrlJ:
		return s, ActionDraw, RenderRequest{Query: s.Query, ModeIndicator: " I ", NavDelta: 1}
	case tcell.KeyCtrlD:
		return s, ActionDraw, RenderRequest{Query: s.Query, ModeIndicator: " I ", NavPage: 1}
	case tcell.KeyCtrlU:
		return s, ActionDraw, RenderRequest{Query: s.Query, ModeIndicator: " I ", NavPage: -1}
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(s.Query) > 0 {
			s.Query = s.Query[:len(s.Query)-1]
		}
		return s, ActionDraw, RenderRequest{Query: s.Query, ModeIndicator: " I "}
	case tcell.KeyEsc:
		if len(s.Query) > 0 {
			s.Query = ""
			return s, ActionDraw, RenderRequest{Query: s.Query, ModeIndicator: " I "}
		}
		return NormalState{}, ActionDraw, RenderRequest{ModeIndicator: " N "}
	case tcell.KeyEnter:
		return s, ActionSelect, RenderRequest{Query: s.Query, ModeIndicator: " I "}
	case tcell.KeyCtrlC:
		return s, ActionQuit, RenderRequest{}
	default:
		if e.Rune >= 32 && e.Rune < 127 {
			s.Query += string(e.Rune)
		}
		return s, ActionDraw, RenderRequest{Query: s.Query, ModeIndicator: " I "}
	}
}

type NormalState struct {
	Count int
}

func (s NormalState) HandleKey(e KeyEvent) (State, Action, RenderRequest) {
	switch e.Key {
	case tcell.KeyUp, tcell.KeyCtrlP, tcell.KeyCtrlK:
		return s, ActionDraw, RenderRequest{ModeIndicator: " N ", NavDelta: -1}
	case tcell.KeyDown, tcell.KeyCtrlN, tcell.KeyCtrlJ:
		return s, ActionDraw, RenderRequest{ModeIndicator: " N ", NavDelta: 1}
	case tcell.KeyCtrlD:
		delta := s.Count
		if delta == 0 {
			delta = 1
		}
		s.Count = 0
		return s, ActionDraw, RenderRequest{ModeIndicator: " N ", NavPage: delta}
	case tcell.KeyCtrlU:
		delta := s.Count
		if delta == 0 {
			delta = 1
		}
		s.Count = 0
		return s, ActionDraw, RenderRequest{ModeIndicator: " N ", NavPage: -delta}
	case tcell.KeyEsc:
		s.Count = 0
		return ConfirmQuitState{}, ActionDraw, RenderRequest{ModeIndicator: " ? "}
	case tcell.KeyEnter:
		s.Count = 0
		return s, ActionSelect, RenderRequest{ModeIndicator: " N "}
	case tcell.KeyCtrlC:
		s.Count = 0
		return s, ActionQuit, RenderRequest{}
	}
	switch e.Rune {
	case 'j', 'J':
		delta := s.Count
		if delta == 0 {
			delta = 1
		}
		s.Count = 0
		return s, ActionDraw, RenderRequest{ModeIndicator: " N ", NavDelta: delta}
	case 'k', 'K':
		delta := s.Count
		if delta == 0 {
			delta = -1
		} else {
			delta = -delta
		}
		s.Count = 0
		return s, ActionDraw, RenderRequest{ModeIndicator: " N ", NavDelta: delta}
	case 'g':
		if s.Count > 0 {
			idx := s.Count - 1
			s.Count = 0
			return s, ActionDraw, RenderRequest{ModeIndicator: " N ", NavTarget: idx}
		}
		s.Count = 0
		return s, ActionDraw, RenderRequest{ModeIndicator: " N ", NavTop: true}
	case 'G':
		if s.Count > 0 {
			idx := s.Count - 1
			s.Count = 0
			return s, ActionDraw, RenderRequest{ModeIndicator: " N ", NavTarget: idx}
		}
		s.Count = 0
		return s, ActionDraw, RenderRequest{ModeIndicator: " N ", NavBottom: true}
	case 'i':
		s.Count = 0
		return InsertState{Query: ""}, ActionDraw, RenderRequest{ModeIndicator: " I ", Query: ""}
	case 'q', 'Q':
		s.Count = 0
		return ConfirmQuitState{}, ActionDraw, RenderRequest{ModeIndicator: " ? "}
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		s.Count = s.Count*10 + int(e.Rune-'0')
		return s, ActionDraw, RenderRequest{ModeIndicator: " N "}
	}
	s.Count = 0
	return s, ActionDraw, RenderRequest{ModeIndicator: " N "}
}

type ConfirmQuitState struct{}

func (s ConfirmQuitState) HandleKey(e KeyEvent) (State, Action, RenderRequest) {
	switch e.Key {
	case tcell.KeyEsc:
		return NormalState{}, ActionDraw, RenderRequest{ModeIndicator: " N "}
	case tcell.KeyEnter:
		return s, ActionQuit, RenderRequest{ModeIndicator: " ? "}
	case tcell.KeyCtrlC:
		return s, ActionQuit, RenderRequest{}
	}
	switch e.Rune {
	case 'y', 'Y':
		return s, ActionQuit, RenderRequest{ModeIndicator: " ? "}
	case 'n', 'N':
		return NormalState{}, ActionDraw, RenderRequest{ModeIndicator: " N "}
	}
	return s, ActionDraw, RenderRequest{ModeIndicator: " ? "}
}
