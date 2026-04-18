package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/ktr0731/go-fuzzyfinder/matching"

	"gwt/internal/git"
)

type TUIColors struct {
	SelectedBg     tcell.Color
	SelectedFg     tcell.Color
	ColorNormal    tcell.Color
	ColorInsert    tcell.Color
	ColorQuit      tcell.Color
	FilterBg       tcell.Color
	FilterFg       tcell.Color
	ModeFg         tcell.Color
	HighlightSelFg tcell.Color
	HighlightFg    tcell.Color
	ArrowFg        tcell.Color
	WindowRatio    float64
}

func DefaultTUIColors() TUIColors {
	return TUIColors{
		SelectedBg:     tcell.NewRGBColor(73, 77, 100),
		SelectedFg:     tcell.NewRGBColor(249, 226, 175),
		ColorNormal:    tcell.ColorBlue,
		ColorInsert:    tcell.ColorGreen,
		ColorQuit:      tcell.ColorRed,
		FilterBg:       tcell.ColorBlack,
		FilterFg:       tcell.ColorWhite,
		ModeFg:         tcell.ColorBlack,
		HighlightSelFg: tcell.NewRGBColor(107, 200, 122),
		HighlightFg:    tcell.NewRGBColor(166, 227, 161),
		ArrowFg:        tcell.ColorRed,
		WindowRatio:    0.75,
	}
}

type matchedWorktree struct {
	WorktreeInfo
	MatchPositions [][2]int
}

func FuzzyMatchPositions(query, target string) [][2]int {
	if query == "" {
		return nil
	}
	queryLower := strings.ToLower(query)
	targetLower := strings.ToLower(target)
	var positions [][2]int
	targetRunes := []rune(targetLower)
	targetIdx := 0
	for _, q := range queryLower {
		found := false
		for ; targetIdx < len(targetRunes); targetIdx++ {
			if targetRunes[targetIdx] == q {
				positions = append(positions, [2]int{targetIdx, targetIdx + 1})
				targetIdx++
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}
	return positions
}

func SelectWorktreeTUI(
	worktrees []WorktreeInfo,
	commander git.Commander,
	colors TUIColors,
) (int, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return -1, err
	}
	defer func() {
		if screen != nil {
			screen.Fini()
		}
	}()

	if err := screen.Init(); err != nil {
		screen = nil
		return -1, err
	}

	var state State = InsertState{Query: ""}

	commonDir, err := commander.GitCommonDir()
	rootIdx := -1
	if err == nil && len(worktrees) > 0 {
		for i, wt := range worktrees {
			gitDir := wt.AbsPath + "/.git"
			if gitDir == commonDir || strings.HasPrefix(commonDir, gitDir+"/") {
				rootIdx = i
				break
			}
		}
	}

	selected := 0
	windowTop := 0
	for i, wt := range worktrees {
		if wt.IsCwd {
			selected = i
			break
		}
	}

	if rootIdx >= 0 && selected == rootIdx && len(worktrees) > 1 {
		selected = len(worktrees) - 1
	} else if rootIdx >= 0 && selected != rootIdx {
		selected = rootIdx
	}

	if selected >= 0 {
		_, screenHeight := screen.Size()
		listHeight := screenHeight - 1
		windowHeight := int(float64(listHeight) * colors.WindowRatio)
		if windowHeight > 1 && selected >= windowHeight-1 {
			windowTop = selected - windowHeight + 1
		}
	}

	numDigits := len(fmt.Sprintf("%d", len(worktrees)))
	numPad := fmt.Sprintf("%%%dd", numDigits)

	displayWorktrees := func(query string) []matchedWorktree {
		if query == "" {
			result := make([]matchedWorktree, len(worktrees))
			for i, wt := range worktrees {
				result[i] = matchedWorktree{WorktreeInfo: wt}
			}
			return result
		}

		displayStrs := make([]string, len(worktrees))
		for i, wt := range worktrees {
			displayStrs[i] = wt.Display
		}

		matched := matching.FindAll(query, displayStrs, matching.WithMode(matching.ModeSmart))
		if len(matched) == 0 {
			return []matchedWorktree{}
		}

		result := make([]matchedWorktree, 0, len(matched))
		for i := len(matched) - 1; i >= 0; i-- {
			m := matched[i]
			wt := worktrees[m.Idx]
			positions := FuzzyMatchPositions(query, wt.Display)
			result = append(result, matchedWorktree{
				WorktreeInfo:   wt,
				MatchPositions: positions,
			})
		}
		return result
	}

	draw := func(req RenderRequest) {
		screen.Clear()
		width, height := screen.Size()

		visible := displayWorktrees(req.Query)

		selectedIdx := selected
		if selectedIdx < 0 {
			selectedIdx = len(visible) - 1
		} else if selectedIdx >= len(visible) {
			selectedIdx = 0
		}

		listHeight := height - 1
		windowHeight := int(float64(listHeight) * colors.WindowRatio)
		if windowHeight < 1 {
			windowHeight = 1
		}
		if windowHeight < len(visible) && selectedIdx >= windowTop+windowHeight {
			windowTop = selectedIdx - windowHeight + 1
		}
		if windowTop < 0 {
			windowTop = 0
		}
		if windowTop+windowHeight > len(visible) {
			windowTop = len(visible) - windowHeight
		}
		if windowTop < 0 {
			windowTop = 0
		}

		arrowStyle := tcell.StyleDefault.Foreground(colors.ArrowFg)
		numberStyle := tcell.StyleDefault.Foreground(tcell.ColorDarkGray)

		for i := windowTop; i < len(visible); i++ {
			wt := visible[i]
			totalShow := windowHeight
			if totalShow > len(visible) {
				totalShow = len(visible)
			}
			startRow := listHeight - totalShow
			row := startRow + (i - windowTop)
			if row >= height-1 || row < 0 {
				continue
			}
			if i == selectedIdx {
				screen.SetContent(0, row, '>', nil, arrowStyle)
			}
			screen.SetContent(1, row, ' ', nil, tcell.StyleDefault)

			numStr := fmt.Sprintf(numPad, i+1)
			for x, r := range numStr {
				if x+2 >= width {
					break
				}
				screen.SetContent(x+2, row, r, nil, numberStyle)
			}
			numberEnd := 2 + numDigits

			style := tcell.StyleDefault
			if i == selectedIdx {
				style = style.Background(colors.SelectedBg).Foreground(colors.SelectedFg).Bold(true)
			}
			for x, r := range wt.Display {
				if x+numberEnd+1 >= width {
					break
				}
				charStyle := style
				for _, pos := range wt.MatchPositions {
					if x >= pos[0] && x < pos[1] {
						if i == selectedIdx {
							charStyle = style.Foreground(colors.HighlightSelFg).Bold(true)
						} else {
							charStyle = style.Foreground(colors.HighlightFg)
						}
						break
					}
				}
				screen.SetContent(x+numberEnd+1, row, r, nil, charStyle)
			}
		}

		modeStyle := tcell.StyleDefault.Foreground(colors.ModeFg)
		switch req.ModeIndicator {
		case " N ":
			modeStyle = modeStyle.Background(colors.ColorNormal)
		case " I ":
			modeStyle = modeStyle.Background(colors.ColorInsert)
		case " ? ":
			modeStyle = modeStyle.Background(colors.ColorQuit)
		}

		leftWidth := 2 + numDigits
		offset := (leftWidth + 1) % 2
		modeRunes := []rune(req.ModeIndicator)
		modeStyle = modeStyle.Bold(true)
		for col := 0; col < leftWidth; col++ {
			idx := col - offset
			if idx < 0 || idx >= len(modeRunes) {
				continue
			}
			screen.SetContent(col, height-1, modeRunes[idx], nil, modeStyle)
		}

		if req.ModeIndicator == " I " {
			for x, r := range req.Query {
				if x+numDigits+3 >= width {
					break
				}
				screen.SetContent(
					x+numDigits+3,
					height-1,
					r,
					nil,
					tcell.StyleDefault.Background(colors.FilterBg).Foreground(colors.FilterFg),
				)
			}
		} else if req.ModeIndicator == " ? " {
			prompt := "[Y/Enter] yes / [N/ESC] no"
			for x, r := range prompt {
				if x+numDigits+3 >= width {
					break
				}
				screen.SetContent(x+numDigits+3, height-1, r, nil, tcell.StyleDefault.Foreground(colors.ColorQuit))
			}
		}

		screen.Show()
	}

	var action Action
	var req RenderRequest
	state, action, req = state.HandleKey(KeyEvent{})
	currentQuery := ""
	draw(req)
	if selected < 0 {
		selected = len(worktrees) - 1
	}

	for {
		ev := screen.PollEvent()

		var keyEvent KeyEvent
		if ev, ok := ev.(*tcell.EventKey); ok {
			keyEvent = KeyEvent{Key: ev.Key(), Rune: ev.Rune()}
		}

		var nextState State
		nextState, action, req = state.HandleKey(keyEvent)

		if req.NavDelta != 0 || req.NavTop || req.NavBottom || req.NavTarget > 0 ||
			req.NavPage != 0 {
			visible := displayWorktrees(currentQuery)
			if req.NavTop {
				selected = 0
			} else if req.NavBottom {
				if len(visible) > 0 {
					selected = len(visible) - 1
				} else {
					selected = 0
				}
			} else if req.NavTarget > 0 {
				targetIdx := req.NavTarget
				if targetIdx < 0 {
					targetIdx = 0
				}
				if targetIdx >= len(visible) {
					targetIdx = len(visible) - 1
				}
				selected = targetIdx
			} else if req.NavPage != 0 {
				_, screenHeight := screen.Size()
				listHeight := screenHeight - 1
				windowHeight := int(float64(listHeight) * colors.WindowRatio)
				if windowHeight < 1 {
					windowHeight = 1
				}
				halfPage := windowHeight / 2
				if halfPage < 1 {
					halfPage = 1
				}
				pageDelta := halfPage * req.NavPage
				newSelected := selected + pageDelta
				if newSelected >= len(visible) {
					selected = len(visible) - 1
				} else if newSelected < 0 {
					selected = 0
				} else {
					selected = newSelected
				}
			} else {
				newSelected := selected + req.NavDelta
				if req.NavDelta > 0 {
					if newSelected >= len(visible) {
						selected = len(visible) - 1
					} else {
						selected = newSelected
					}
				} else {
					if newSelected < 0 {
						selected = 0
					} else {
						selected = newSelected
					}
				}
			}

			_, screenHeight := screen.Size()
			listHeight := screenHeight - 1
			windowHeight := int(float64(listHeight) * colors.WindowRatio)
			if windowHeight < 1 {
				windowHeight = 1
			}
			if selected < windowTop {
				windowTop = selected
			} else if selected >= windowTop+windowHeight {
				windowTop = selected - windowHeight + 1
			}
		}

		if currentQuery != req.Query {
			oldVisible := displayWorktrees(currentQuery)
			newVisible := displayWorktrees(req.Query)

			var selectedPath string
			if len(oldVisible) > 0 {
				clamped := selected
				if clamped >= len(oldVisible) {
					clamped = len(oldVisible) - 1
				}
				if clamped < 0 {
					clamped = 0
				}
				selectedPath = oldVisible[clamped].AbsPath
			}

			if len(newVisible) > 0 {
				if req.Query == "" {
					selected = 0
					for i, wt := range newVisible {
						if wt.AbsPath == selectedPath {
							selected = i
							break
						}
					}
				} else {
					selected = len(newVisible) - 1
				}
			} else {
				selected = 0
			}
		}

		currentQuery = req.Query
		state = nextState
		draw(req)

		if action == ActionQuit {
			return -1, nil
		}
		if action == ActionSelect {
			visible := displayWorktrees(currentQuery)
			selectedIdx := selected
			if selectedIdx >= len(visible) {
				selectedIdx = len(visible) - 1
			}
			if selectedIdx < 0 {
				selectedIdx = 0
			}
			if selectedIdx >= 0 && selectedIdx < len(visible) {
				for i, wt := range worktrees {
					if wt.AbsPath == visible[selectedIdx].AbsPath {
						return i, nil
					}
				}
			}
			return selected, nil
		}
	}
}
