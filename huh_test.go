package huh

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var pretty = lipgloss.NewStyle().
	Border(lipgloss.NormalBorder()).
	MarginTop(1).
	Padding(1, 3, 1, 2)

func TestForm(t *testing.T) {
	type Taco struct {
		Shell    string
		Base     string
		Toppings []string
	}

	type Order struct {
		Taco         Taco
		Name         string
		Instructions string
		Discount     bool
	}

	var taco Taco
	order := Order{Taco: taco}

	f := NewForm(
		NewGroup(
			NewSelect[string]().
				Options(NewOptions("Soft", "Hard")...).
				Title("Shell?").
				Description("Our tortillas are made fresh in-house every day.").
				Validate(func(t string) error {
					if t == "Hard" {
						return fmt.Errorf("we're out of hard shells, sorry")
					}
					return nil
				}).
				Value(&order.Taco.Shell),

			NewSelect[string]().
				Options(NewOptions("Chicken", "Beef", "Fish", "Beans")...).
				Value(&order.Taco.Base).
				Title("Base"),
		),

		// Prompt for toppings and special instructions.
		// The customer can ask for up to 4 toppings.
		NewGroup(
			NewMultiSelect[string]().
				Title("Toppings").
				Description("Choose up to 4.").
				Options(
					NewOption("Lettuce", "lettuce").Selected(true),
					NewOption("Tomatoes", "tomatoes").Selected(true),
					NewOption("Corn", "corn"),
					NewOption("Salsa", "salsa"),
					NewOption("Sour Cream", "sour cream"),
					NewOption("Cheese", "cheese"),
				).
				Validate(func(t []string) error {
					if len(t) <= 0 {
						return fmt.Errorf("at least one topping is required")
					}
					return nil
				}).
				Value(&order.Taco.Toppings).
				Filterable(true).
				Limit(4),
		),

		// Gather final details for the order.
		NewGroup(
			NewInput().
				Value(&order.Name).
				Title("What's your name?").
				Placeholder("Margaret Thatcher").
				Description("For when your order is ready."),

			NewText().
				Value(&order.Instructions).
				Placeholder("Just put it in the mailbox please").
				Title("Special Instructions").
				Description("Anything we should know?").
				CharLimit(400),

			NewConfirm().
				Title("Would you like 15% off?").
				Value(&order.Discount).
				Affirmative("Yes!").
				Negative("No."),
		),
	)

	f.Update(f.Init())

	view := f.View()

	//
	//  ┃ Shell?
	//  ┃ Our tortillas are made fresh in-house every day.
	//  ┃ > Soft
	//  ┃   Hard
	//
	//    Base
	//    > Chicken
	//      Beef
	//      Fish
	//      Beans
	//
	//   ↑ up • ↓ down • / filter • enter select • shift+tab back
	//

	if !strings.Contains(view, "Shell?") {
		t.Log(pretty.Render(view))
		t.Error("Expected form to contain Shell? title")
	}

	if !strings.Contains(view, "Our tortillas are made fresh in-house every day.") {
		t.Log(pretty.Render(view))
		t.Error("Expected form to contain tortilla description")
	}

	if !strings.Contains(view, "Base") {
		t.Log(pretty.Render(view))
		t.Error("Expected form to contain Base title")
	}

	// Attempt to select hard shell and retrieve error.
	m, _ := f.Update(keys('j'))
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	view = m.View()

	if !strings.Contains(view, "* we're out of hard shells, sorry") {
		t.Log(pretty.Render(view))
		t.Error("Expected form to show out of hard shells error")
	}

	m, _ = m.Update(keys('k'))

	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = batchUpdate(m, cmd)

	view = m.View()

	if !strings.Contains(view, "┃ > Chicken") {
		t.Log(pretty.Render(view))
		t.Fatal("Expected form to continue to base group")
	}

	// batchMsg + nextGroup
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = batchUpdate(m, cmd)
	view = m.View()

	//
	// ┃ Toppings
	// ┃ Choose up to 4.
	// ┃ > ✓ Lettuce
	// ┃   ✓ Tomatoes
	// ┃   • Corn
	// ┃   • Salsa
	// ┃   • Sour Cream
	// ┃   • Cheese
	//
	//  x toggle • ↑ up • ↓ down • enter confirm • shift+tab back
	//
	if !strings.Contains(view, "Toppings") {
		t.Log(pretty.Render(view))
		t.Fatal("Expected form to show toppings group")
	}

	if !strings.Contains(view, "Choose up to 4.") {
		t.Log(pretty.Render(view))
		t.Error("Expected form to show toppings description")
	}

	if !strings.Contains(view, "> ✓ Lettuce ") {
		t.Log(pretty.Render(view))
		t.Error("Expected form to preselect lettuce")
	}

	if !strings.Contains(view, "  ✓ Tomatoes") {
		t.Log(pretty.Render(view))
		t.Error("Expected form to preselect tomatoes")
	}

	m, _ = m.Update(keys('j'))
	m, _ = m.Update(keys('j'))
	view = m.View()

	if !strings.Contains(view, "> • Corn") {
		t.Log(pretty.Render(view))
		t.Error("Expected form to change selection to corn")
	}

	m, _ = m.Update(keys('x'))
	view = m.View()

	if !strings.Contains(view, "> ✓ Corn") {
		t.Log(pretty.Render(view))
		t.Error("Expected form to change selection to corn")
	}

	m = batchUpdate(m.Update(tea.KeyMsg{Type: tea.KeyEnter}))
	view = m.View()

	if !strings.Contains(view, "What's your name?") {
		t.Log(pretty.Render(view))
		t.Error("Expected form to prompt for name")
	}

	if !strings.Contains(view, "Special Instructions") {
		t.Log(pretty.Render(view))
		t.Error("Expected form to prompt for special instructions")
	}

	if !strings.Contains(view, "Would you like 15% off?") {
		t.Log(pretty.Render(view))
		t.Error("Expected form to prompt for discount")
	}

	//
	// ┃ What's your name?
	// ┃ For when your order is ready.
	// ┃ > Margaret Thatcher
	//
	//    Special Instructions
	//    Anything we should know?
	//    Just put it in the mailbox please
	//
	//    Would you like 15% off?
	//
	//      Yes!     No.
	//
	//   enter next • shift+tab back
	//
	m.Update(keys('G', 'l', 'e', 'n'))
	view = m.View()
	if !strings.Contains(view, "Glen") {
		t.Log(pretty.Render(view))
		t.Error("Expected form to accept user input")
	}

	if order.Taco.Shell != "Soft" {
		t.Error("Expected order shell to be Soft")
	}

	if order.Taco.Base != "Chicken" {
		t.Error("Expected order shell to be Chicken")
	}

	if len(order.Taco.Toppings) != 3 {
		t.Error("Expected order to have 3 toppings")
	}

	if order.Name != "Glen" {
		t.Error("Expected order name to be Glen")
	}

	// TODO: Finish and submit form.
}

func TestInput(t *testing.T) {
	field := NewInput()
	f := NewForm(NewGroup(field))
	f.Update(f.Init())

	view := f.View()

	if !strings.Contains(view, ">") {
		t.Log(pretty.Render(view))
		t.Error("Expected field to contain prompt.")
	}

	// Type Huh in the form.
	m, _ := f.Update(keys('H', 'u', 'h'))
	f = m.(*Form)
	view = f.View()

	if !strings.Contains(view, "Huh") {
		t.Log(pretty.Render(view))
		t.Error("Expected field to contain Huh.")
	}

	if !strings.Contains(view, "enter next • shift+tab back") {
		t.Log(pretty.Render(view))
		t.Error("Expected field to contain help.")
	}
}

func TestText(t *testing.T) {
	field := NewText()
	f := NewForm(NewGroup(field))
	f.Update(f.Init())

	// Type Huh in the form.
	m, _ := f.Update(keys('H', 'u', 'h'))
	f = m.(*Form)
	view := f.View()

	if !strings.Contains(view, "Huh") {
		t.Log(pretty.Render(view))
		t.Error("Expected field to contain Huh.")
	}

	if !strings.Contains(view, "enter next • alt+enter / ctrl+j new line • ctrl+e open editor • shift+tab back") {
		t.Log(pretty.Render(view))
		t.Error("Expected field to contain help.")
	}
}

func TestConfirm(t *testing.T) {
	field := NewConfirm().Title("Are you sure?")
	f := NewForm(NewGroup(field))
	f.Update(f.Init())

	// Type Huh in the form.
	m, _ := f.Update(keys('H'))
	f = m.(*Form)
	view := f.View()

	if !strings.Contains(view, "Yes") {
		t.Log(pretty.Render(view))
		t.Error("Expected field to contain Yes.")
	}

	if !strings.Contains(view, "No") {
		t.Log(pretty.Render(view))
		t.Error("Expected field to contain No.")
	}

	if !strings.Contains(view, "Are you sure?") {
		t.Log(pretty.Render(view))
		t.Error("Expected field to contain Are you sure?.")
	}

	if !strings.Contains(view, "←/→ toggle • enter next • shift+tab back") {
		t.Log(pretty.Render(view))
		t.Error("Expected field to contain help.")
	}
}

func TestSelect(t *testing.T) {
	field := NewSelect[string]().Options(NewOptions("Foo", "Bar", "Baz")...).Title("Which one?")
	f := NewForm(NewGroup(field))
	f.Update(f.Init())

	view := f.View()

	if !strings.Contains(view, "Foo") {
		t.Log(pretty.Render(view))
		t.Error("Expected field to contain Foo.")
	}

	if !strings.Contains(view, "Which one?") {
		t.Log(pretty.Render(view))
		t.Error("Expected field to contain Which one?.")
	}

	// Move selection cursor down
	if !strings.Contains(view, "> Foo") {
		t.Log(pretty.Render(view))
		t.Error("Expected cursor to be on Foo.")
	}

	m, _ := f.Update(keys('j'))
	f = m.(*Form)

	view = f.View()

	if strings.Contains(view, "> Foo") {
		t.Log(pretty.Render(view))
		t.Error("Expected cursor to be on Bar.")
	}

	if !strings.Contains(view, "> Bar") {
		t.Log(pretty.Render(view))
		t.Error("Expected cursor to be on Bar.")
	}

	if !strings.Contains(view, "↑ up • ↓ down • / filter • enter select • shift+tab back") {
		t.Log(pretty.Render(view))
		t.Error("Expected field to contain help.")
	}
}

func TestMultiSelect(t *testing.T) {
	field := NewMultiSelect[string]().Options(NewOptions("Foo", "Bar", "Baz")...).Title("Which one?")
	f := NewForm(NewGroup(field))
	f.Update(f.Init())

	view := f.View()

	if !strings.Contains(view, "Foo") {
		t.Log(pretty.Render(view))
		t.Error("Expected field to contain Foo.")
	}

	if !strings.Contains(view, "Which one?") {
		t.Log(pretty.Render(view))
		t.Error("Expected field to contain Which one?.")
	}

	if !strings.Contains(view, "> • Foo") {
		t.Log(pretty.Render(view))
		t.Error("Expected cursor to be on Foo.")
	}

	// Move selection cursor down
	m, _ := f.Update(keys('j'))
	view = m.View()

	if strings.Contains(view, "> • Foo") {
		t.Log(pretty.Render(view))
		t.Error("Expected cursor to be on Bar.")
	}

	if !strings.Contains(view, "> • Bar") {
		t.Log(pretty.Render(view))
		t.Error("Expected cursor to be on Bar.")
	}

	// Toggle
	m, _ = f.Update(keys('x'))
	view = m.View()

	if !strings.Contains(view, "> ✓ Bar") {
		t.Log(pretty.Render(view))
		t.Error("Expected cursor to be on Bar.")
	}

	if !strings.Contains(view, "x toggle • ↑ up • ↓ down • enter confirm • shift+tab back") {
		t.Log(pretty.Render(view))
		t.Error("Expected field to contain help.")
	}
}

func TestHideGroup(t *testing.T) {
	f := NewForm(
		NewGroup(NewNote().Description("Foo")).WithHide(true),
		NewGroup(NewNote().Description("Bar")),
		NewGroup(NewNote().Description("Baz")),
		NewGroup(NewNote().Description("Qux")).WithHideFunc(func() bool { return false }).WithHide(true),
	)

	f = batchUpdate(f, f.Init()).(*Form)

	if v := f.View(); !strings.Contains(v, "Bar") {
		t.Log(pretty.Render(v))
		t.Error("expected Bar to not be hidden")
	}

	f.Update(nextGroup())

	if v := f.View(); !strings.Contains(v, "Baz") {
		t.Log(pretty.Render(v))
		t.Error("expected Baz to not be hidden")
	}

	f = batchUpdate(f, nextGroup).(*Form)

	if v := f.View(); strings.Contains(v, "Qux") {
		t.Log(pretty.Render(v))
		t.Error("expected Qux to be hidden")
	}
}

func TestNote(t *testing.T) {
	field := NewNote().Title("Taco").Description("How may we take your order?").Next(true)
	f := NewForm(NewGroup(field))
	f.theme.Focused.Base.Border(lipgloss.HiddenBorder())
	f.Update(f.Init())

	view := f.View()

	if !strings.Contains(view, "Taco") {
		t.Log(view)
		t.Error("Expected field to contain Taco title.")
	}

	if !strings.Contains(view, "order?") {
		t.Log(view)
		t.Error("Expected field to contain Taco description.")
	}

	if !strings.Contains(view, "Next") {
		t.Log(view)
		t.Error("Expected field to contain next button")
	}

	if !strings.Contains(view, "enter next") {
		t.Log(view)
		t.Error("Expected field to contain help.")
	}
}

func batchUpdate(m tea.Model, cmd tea.Cmd) tea.Model {
	if cmd == nil {
		return m
	}

	msg := cmd()
	if msg == nil {
		return m
	}

	switch msg := msg.(type) {
	case tea.BatchMsg:
		for _, c := range msg {
			m, cmd = m.Update(c())
			if cmd == nil {
				continue
			}
		}
		return batchUpdate(m, cmd)
	}

	m, cmd = m.Update(msg)
	return batchUpdate(m, cmd)
}

func keys(runes ...rune) tea.KeyMsg {
	return tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: runes,
	}
}
