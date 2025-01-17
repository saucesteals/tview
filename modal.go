package tview

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

// Modal is a centered message window used to inform the user or prompt them
// for an immediate decision. It needs to have at least one button (added via
// AddButtons()) or it will never disappear.
//
// See https://github.com/saucesteals/tview/wiki/Modal for an example.
type Modal struct {
	*Box

	// The Frame embedded in the modal.
	Frame *Frame

	// The Form embedded in the modal's frame.
	Form *Form

	// The message text (original, not word-wrapped).
	text string

	// The text color.
	textColor tcell.Color

	// The optional callback for when the user clicked one of the buttons. It
	// receives the index of the clicked button and the button's label.
	done func(buttonIndex int, buttonLabel string)
}

// NewModal returns a new modal message window.
func NewModal() *Modal {
	m := &Modal{
		Box:       NewBox(),
		textColor: Styles.PrimaryTextColor,
	}
	m.Form = NewForm().
		SetButtonsAlign(AlignCenter).
		SetButtonBackgroundColor(Styles.PrimitiveBackgroundColor).
		SetButtonTextColor(Styles.PrimaryTextColor)
	m.Form.SetBackgroundColor(Styles.ContrastBackgroundColor).SetBorderPadding(0, 0, 0, 0)
	m.Form.SetCancelFunc(func() {
		if m.done != nil {
			m.done(-1, "")
		}
	})
	m.Frame = NewFrame(m.Form).SetBorders(0, 0, 1, 0, 0, 0)
	m.Frame.SetBorder(true).
		SetBackgroundColor(Styles.ContrastBackgroundColor).
		SetBorderPadding(1, 1, 1, 1)
	return m
}

// SetBorderColor sets the color of the border.
func (m *Modal) SetBorderColor(color tcell.Color) *Modal {
	m.Frame.SetBorderColor(color)
	return m
}

// SetBorder enables or disables the display of the border
func (m *Modal) SetBorder(show bool) *Modal {
	m.Frame.SetBorder(show)
	return m
}

// SetBackgroundColor sets the color of the modal frame background.
func (m *Modal) SetBackgroundColor(color tcell.Color) *Modal {
	m.Form.SetBackgroundColor(color)
	m.Frame.SetBackgroundColor(color)
	return m
}

// SetTextColor sets the color of the message text.
func (m *Modal) SetTextColor(color tcell.Color) *Modal {
	m.textColor = color
	return m
}

// SetButtonBackgroundColor sets the background color of the buttons.
func (m *Modal) SetButtonBackgroundColor(color tcell.Color) *Modal {
	m.Form.SetButtonBackgroundColor(color)
	return m
}

// SetButtonTextColor sets the color of the button texts.
func (m *Modal) SetButtonTextColor(color tcell.Color) *Modal {
	m.Form.SetButtonTextColor(color)
	return m
}

// SetDoneFunc sets a handler which is called when one of the buttons was
// pressed. It receives the index of the button as well as its label text. The
// handler is also called when the user presses the Escape key. The index will
// then be negative and the label text an emptry string.
func (m *Modal) SetDoneFunc(handler func(buttonIndex int, buttonLabel string)) *Modal {
	m.done = handler
	return m
}

// SetText sets the message text of the window. The text may contain line
// breaks but color tag states will not transfer to following lines. Note that
// words are wrapped, too, based on the final size of the window.
func (m *Modal) SetText(text string) *Modal {
	m.text = text
	return m
}

// AddButtons adds buttons to the window. There must be at least one button and
// a "done" handler so the window can be closed again.
func (m *Modal) AddButtons(labels []string) *Modal {
	for index, label := range labels {
		func(i int, l string) {
			m.Form.AddButton(label, func() {
				if m.done != nil {
					m.done(i, l)
				}
			})
			button := m.Form.GetButton(m.Form.GetButtonCount() - 1)
			button.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				switch event.Key() {
				case tcell.KeyDown, tcell.KeyRight:
					return tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
				case tcell.KeyUp, tcell.KeyLeft:
					return tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModNone)
				}
				return event
			})
		}(index, label)
	}
	return m
}

// ClearButtons removes all buttons from the window.
func (m *Modal) ClearButtons() *Modal {
	m.Form.ClearButtons()
	return m
}

// SetFocus shifts the focus to the button with the given index.
func (m *Modal) SetFocus(index int) *Modal {
	m.Form.SetFocus(index)
	return m
}

// Focus is called when this primitive receives focus.
func (m *Modal) Focus(delegate func(p Primitive)) {
	delegate(m.Form)
}

// HasFocus returns whether or not this primitive has focus.
func (m *Modal) HasFocus() bool {
	return m.Form.HasFocus()
}

// Draw draws this primitive onto the screen.
func (m *Modal) Draw(screen tcell.Screen) {
	// Calculate the width of this modal.
	buttonsWidth := 0
	for _, button := range m.Form.buttons {
		buttonsWidth += TaggedStringWidth(button.label) + 4 + 2
	}
	buttonsWidth -= 2
	screenWidth, screenHeight := screen.Size()
	width := screenWidth / 3
	if width < buttonsWidth {
		width = buttonsWidth
	}
	// width is now without the box border.

	// Reset the text and find out how wide it is.
	m.Frame.Clear()
	var lines []string
	for _, line := range strings.Split(m.text, "\n") {
		if len(line) == 0 {
			lines = append(lines, "")
			continue
		}
		lines = append(lines, WordWrap(line, width)...)
	}
	//lines := WordWrap(m.text, width)
	for _, line := range lines {
		m.Frame.AddText(line, true, AlignCenter, m.textColor)
	}

	// Set the modal's position and size.
	height := len(lines) + 6
	width += 4
	x := (screenWidth - width) / 2
	y := (screenHeight - height) / 2
	m.SetRect(x, y, width, height)

	// Draw the frame.
	m.Frame.SetRect(x, y, width, height)
	m.Frame.Draw(screen)
}

// MouseHandler returns the mouse handler for this primitive.
func (m *Modal) MouseHandler() func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
	return m.WrapMouseHandler(func(action MouseAction, event *tcell.EventMouse, setFocus func(p Primitive)) (consumed bool, capture Primitive) {
		// Pass mouse events on to the form.
		consumed, capture = m.Form.MouseHandler()(action, event, setFocus)
		if !consumed && action == MouseLeftClick && m.InRect(event.Position()) {
			setFocus(m)
			consumed = true
		}
		return
	})
}

// InputHandler returns the handler for this primitive.
func (m *Modal) InputHandler() func(event *tcell.EventKey, setFocus func(p Primitive)) {
	return m.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p Primitive)) {
		if m.Frame.HasFocus() {
			if handler := m.Frame.InputHandler(); handler != nil {
				handler(event, setFocus)
				return
			}
		}
	})
}
