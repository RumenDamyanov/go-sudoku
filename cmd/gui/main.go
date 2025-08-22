//go:build gui

package main

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"go.rumenx.com/sudoku"
)

// shared state for the GUI grid
type gridState struct {
	size, boxR, boxC int
	entries          [][]*widget.Entry
	bgs              [][]*canvas.Rectangle
	grid             *fyne.Container
	timerStart       time.Time
	timerStop        chan struct{}
	timerLabel       *widget.Label
}

func main() {
	a := app.NewWithID("go.rumenx.com/sudoku/gui")
	a.Settings().SetTheme(newModernTheme())
	w := a.NewWindow("Sudoku — go.rumenx.com/sudoku")
	w.Resize(fyne.NewSize(560, 680))

	// State
	st := &gridState{size: 9, boxR: 3, boxC: 3}
	var toolbar *fyne.Container
	var footer *fyne.Container

	// Builders
	rebuild := func() {
		// recreate entries and backgrounds
		st.entries = make([][]*widget.Entry, st.size)
		st.bgs = make([][]*canvas.Rectangle, st.size)
		grid := container.NewGridWithColumns(st.size)
		for r := 0; r < st.size; r++ {
			st.entries[r] = make([]*widget.Entry, st.size)
			st.bgs[r] = make([]*canvas.Rectangle, st.size)
			for c := 0; c < st.size; c++ {
				// background with alternating sub-box colour
				base := color.NRGBA{R: 245, G: 247, B: 250, A: 255} // light
				alt := color.NRGBA{R: 230, G: 235, B: 240, A: 255}
				if ((r/st.boxR)+(c/st.boxC))%2 == 1 {
					base = alt
				}
				bg := canvas.NewRectangle(base)
				bg.SetMinSize(fyne.NewSize(36, 36))

				e := widget.NewEntry()
				e.SetPlaceHolder("0")
				e.TextStyle = fyne.TextStyle{Monospace: true}
				e.MultiLine = false
				maxDigit := st.size
				e.Validator = func(s string) error {
					if s == "" {
						return nil
					}
					if len(s) > 1 {
						return fmt.Errorf("1 digit")
					}
					if s[0] < '0' || s[0] > '9' {
						return fmt.Errorf("digit 0-9")
					}
					v := int(s[0] - '0')
					if v > maxDigit {
						return fmt.Errorf("max %d", maxDigit)
					}
					return nil
				}

				st.entries[r][c] = e
				st.bgs[r][c] = bg
				grid.Add(container.NewMax(bg, e))
			}
		}
		st.grid = grid
	}

	// Controls
	sizeSelect := widget.NewSelect([]string{"4x4 (2x2)", "6x6 (2x3)", "9x9 (3x3)"}, func(s string) {
		switch {
		case strings.HasPrefix(s, "4x4"):
			st.size, st.boxR, st.boxC = 4, 2, 2
		case strings.HasPrefix(s, "6x6"):
			st.size, st.boxR, st.boxC = 6, 2, 3
		default:
			st.size, st.boxR, st.boxC = 9, 3, 3
		}
		rebuild()
		content := container.NewBorder(toolbar, footer, nil, nil, st.grid)
		w.SetContent(content)
	})
	sizeSelect.Selected = "9x9 (3x3)"

	difficulty := widget.NewRadioGroup([]string{string(sudoku.Easy), string(sudoku.Medium), string(sudoku.Hard)}, nil)
	// Ensure labels render with theme foreground color
	difficulty.Horizontal = true
	difficulty.SetSelected(string(sudoku.Medium))

	// Timer
	st.timerLabel = widget.NewLabel("Time 00:00")
	startTimer := func() {
		if st.timerStop != nil {
			close(st.timerStop)
		}
		st.timerStop = make(chan struct{})
		st.timerStart = time.Now()
		go func(ch <-chan struct{}) {
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					d := time.Since(st.timerStart).Round(time.Second)
					m := int(d.Minutes())
					s := int(d.Seconds()) % 60
					st.timerLabel.SetText(fmt.Sprintf("Time %02d:%02d", m, s))
				case <-ch:
					return
				}
			}
		}(st.timerStop)
	}
	stopTimer := func() {
		if st.timerStop != nil {
			close(st.timerStop)
			st.timerStop = nil
		}
	}

	highlightSelected := func() {
		// Reset all to base, highlight focused cell
		var focused *widget.Entry
		if f := w.Canvas().Focused(); f != nil {
			if e, ok := f.(*widget.Entry); ok {
				focused = e
			}
		}
		for r := 0; r < st.size; r++ {
			for c := 0; c < st.size; c++ {
				// base colour
				base := color.NRGBA{R: 245, G: 247, B: 250, A: 255}
				alt := color.NRGBA{R: 230, G: 235, B: 240, A: 255}
				if ((r/st.boxR)+(c/st.boxC))%2 == 1 {
					base = alt
				}
				st.bgs[r][c].FillColor = base
				if focused != nil && st.entries[r][c] == focused {
					st.bgs[r][c].FillColor = color.NRGBA{R: 204, G: 231, B: 255, A: 255}
				}
				st.bgs[r][c].Refresh()
			}
		}
	}
	// poll focus changes lightly
	go func() {
		ticker := time.NewTicker(150 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			highlightSelected()
		}
	}()

	btnGenerate := widget.NewButton("Generate", func() {
		var d sudoku.Difficulty
		switch difficulty.Selected { // default to medium
		case string(sudoku.Easy):
			d = sudoku.Easy
		case string(sudoku.Hard):
			d = sudoku.Hard
		default:
			d = sudoku.Medium
		}
		g, _ := sudoku.NewGrid(st.size, st.boxR, st.boxC)
		puz, err := g.Generate(d, 1)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		setGrid(st, puz, true)
		startTimer()
	})

	btnSolve := widget.NewButton("Solve", func() {
		g, err := gridFromEntries(st)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		if sol, ok := g.Solve(); ok {
			setGrid(st, sol, false)
			stopTimer()
		} else {
			dialog.ShowInformation("Unsolvable", "This puzzle has no solution.", w)
		}
	})

	btnValidate := widget.NewButton("Validate", func() {
		g, err := gridFromEntries(st)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		if err := g.Validate(); err != nil {
			dialog.ShowError(fmt.Errorf("invalid: %w", err), w)
		} else {
			dialog.ShowInformation("OK", "Board is valid (no duplicate rows/cols/boxes).", w)
		}
	})

	btnHint := widget.NewButton("Hint", func() {
		g, err := gridFromEntries(st)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		if r, c, v, ok := sudoku.HintGrid(g); ok {
			// if a specific cell is focused and empty, prefer that
			if f := w.Canvas().Focused(); f != nil {
				if fe, okE := f.(*widget.Entry); okE {
					rr, cc := findEntry(st, fe)
					if rr >= 0 && g.Cells[rr][cc] == 0 {
						r, c = rr, cc
						// compute value from solution
						if sol, ok2 := g.Solve(); ok2 {
							v = sol.Cells[r][c]
						}
					}
				}
			}
			st.entries[r][c].SetText(strconv.Itoa(v))
			st.entries[r][c].Enable() // hint is user input
		} else {
			dialog.ShowInformation("No hint", "Board is invalid or solved.", w)
		}
	})

	btnClear := widget.NewButton("Clear", func() {
		g, _ := sudoku.NewGrid(st.size, st.boxR, st.boxC)
		setGrid(st, g, false)
		stopTimer()
		st.timerLabel.SetText("Time 00:00")
	})

	// Toolbar with theme-aware background for good contrast in light/dark modes
	labelSize := widget.NewLabel("Size:")
	labelDiff := widget.NewLabel("Difficulty:")
	labelSize.TextStyle = fyne.TextStyle{Bold: true}
	labelDiff.TextStyle = fyne.TextStyle{Bold: true}
	// Put difficulty group over a subtle contrasting background to improve legibility
	diffBG := canvas.NewRectangle(color.NRGBA{R: 0, G: 0, B: 0, A: 0}) // transparent (theme handles colors)
	diffWrap := container.NewMax(diffBG, container.NewPadded(difficulty))
	tbInner := container.NewHBox(
		labelSize, sizeSelect,
		labelDiff, diffWrap,
		btnGenerate, btnSolve, btnValidate, btnHint, btnClear,
	)
	tbBG := canvas.NewRectangle(theme.BackgroundColor())
	tbBG.SetMinSize(fyne.NewSize(0, 40))
	toolbar = container.NewMax(tbBG, container.NewPadded(tbInner))

	footer = container.NewHBox(widget.NewLabel("Press Hint to fill a cell"), layout.NewSpacer(), st.timerLabel)

	// initial build
	rebuild()
	content := container.NewBorder(toolbar, footer, nil, nil, st.grid)
	w.SetContent(content)
	w.ShowAndRun()
}

func setGrid(st *gridState, g sudoku.Grid, lockNonZero bool) {
	// re-dimension if needed
	if st.size != g.Size || st.boxR != g.BoxRows || st.boxC != g.BoxCols {
		st.size, st.boxR, st.boxC = g.Size, g.BoxRows, g.BoxCols
	}
	for r := 0; r < st.size; r++ {
		for c := 0; c < st.size; c++ {
			v := g.Cells[r][c]
			if v == 0 {
				st.entries[r][c].SetText("")
				st.entries[r][c].Enable()
				continue
			}
			st.entries[r][c].SetText(strconv.Itoa(v))
			if lockNonZero {
				st.entries[r][c].Disable()
			} else {
				st.entries[r][c].Enable()
			}
		}
	}
}

func gridFromEntries(st *gridState) (sudoku.Grid, error) {
	g, _ := sudoku.NewGrid(st.size, st.boxR, st.boxC)
	for r := 0; r < st.size; r++ {
		for c := 0; c < st.size; c++ {
			s := st.entries[r][c].Text
			if s == "" {
				g.Cells[r][c] = 0
				continue
			}
			if len(s) != 1 || s[0] < '0' || s[0] > '9' {
				return sudoku.Grid{}, fmt.Errorf("invalid value at (%d,%d)", r+1, c+1)
			}
			v := int(s[0] - '0')
			if v > st.size {
				return sudoku.Grid{}, fmt.Errorf("value exceeds grid size at (%d,%d)", r+1, c+1)
			}
			g.Cells[r][c] = v
		}
	}
	return g, nil
}

func findEntry(st *gridState, e *widget.Entry) (int, int) {
	for r := 0; r < st.size; r++ {
		for c := 0; c < st.size; c++ {
			if st.entries[r][c] == e {
				return r, c
			}
		}
	}
	return -1, -1
}

// theme
type modernTheme struct{}

func newModernTheme() fyne.Theme { return &modernTheme{} }

func (m *modernTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	// Provide high-contrast palettes for both light and dark variants.
	if v == theme.VariantDark {
		switch n {
		case theme.ColorNameBackground:
			return color.NRGBA{R: 15, G: 23, B: 42, A: 255} // slate-900
		case theme.ColorNameForeground:
			return color.NRGBA{R: 241, G: 245, B: 249, A: 255} // slate-100
		case theme.ColorNameInputBackground:
			return color.NRGBA{R: 30, G: 41, B: 59, A: 255} // slate-800
		case theme.ColorNamePlaceHolder:
			return color.NRGBA{R: 148, G: 163, B: 184, A: 255} // slate-400
		case theme.ColorNameHover:
			return color.NRGBA{R: 51, G: 65, B: 85, A: 255} // slate-700
		case theme.ColorNameFocus:
			return color.NRGBA{R: 96, G: 165, B: 250, A: 255} // blue-400
		case theme.ColorNamePressed:
			return color.NRGBA{R: 30, G: 41, B: 59, A: 255} // slate-800
		case theme.ColorNameHyperlink:
			return color.NRGBA{R: 147, G: 197, B: 253, A: 255} // blue-300
		case theme.ColorNameButton, theme.ColorNamePrimary:
			return color.NRGBA{R: 37, G: 99, B: 235, A: 255} // blue-600
		case theme.ColorNameDisabled:
			return color.NRGBA{R: 107, G: 114, B: 128, A: 255} // gray-500
		}
		return theme.DarkTheme().Color(n, v)
	}
	// Light variant
	switch n {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 250, G: 252, B: 255, A: 255} // #FAFCFF
	case theme.ColorNameForeground:
		return color.NRGBA{R: 15, G: 23, B: 42, A: 255} // slate-900
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 255, G: 255, B: 255, A: 255} // white
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 148, G: 163, B: 184, A: 255} // slate-400
	case theme.ColorNameHover:
		return color.NRGBA{R: 227, G: 242, B: 253, A: 255} // #E3F2FD
	case theme.ColorNameFocus:
		return color.NRGBA{R: 66, G: 133, B: 244, A: 255} // #4285F4
	case theme.ColorNamePressed:
		return color.NRGBA{R: 198, G: 219, B: 252, A: 255} // #C6DBFC
	case theme.ColorNameHyperlink:
		return color.NRGBA{R: 37, G: 99, B: 235, A: 255} // #2563EB
	case theme.ColorNameButton, theme.ColorNamePrimary:
		return color.NRGBA{R: 37, G: 99, B: 235, A: 255} // #2563EB
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 156, G: 163, B: 175, A: 255} // #9CA3AF
	}
	return theme.LightTheme().Color(n, v)
}
func (m *modernTheme) Icon(n fyne.ThemeIconName) fyne.Resource { return theme.DarkTheme().Icon(n) }
func (m *modernTheme) Font(s fyne.TextStyle) fyne.Resource     { return theme.DarkTheme().Font(s) }
func (m *modernTheme) Size(n fyne.ThemeSizeName) float32       { return theme.DarkTheme().Size(n) }
