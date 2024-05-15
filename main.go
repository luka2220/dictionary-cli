package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// initialize model and program options
	m := New()
	p := tea.NewProgram(m)

	// run the cli
	if _, err := p.Run(); err != nil {
		fmt.Printf("An error occurred starting the program: %v", err)
		os.Exit(1)
	}
}

// Model: app state
type Model struct {
	title     string
	terms     Terms
	textinput textinput.Model
	err       error
}

// NewModel: initial model
func New() Model {
	ti := textinput.New()
	ti.Placeholder = "Search word index"
	ti.Focus()

	return Model{
		title:     "Program running",
		textinput: ti,
	}
}

// Init: start the event loop
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// Update: handle messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	// Switch though msg types
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
		switch msg.Type {
		case tea.KeyEnter:
			v := m.textinput.Value()
			return m, handleQuerySearch(v)
		}

	case TermsResponseMessage:
		if msg.Err != nil {
			m.err = msg.Err
		}
		m.terms = msg.Terms

		return m, nil
	}

	// initalize a command
	var cmd tea.Cmd

	m.textinput, cmd = m.textinput.Update(msg)

	return m, cmd
}

// View: return a string based on the view of the model
func (m Model) View() string {
	s := m.textinput.View() + "\n\n" // Get the current state of the text input

	if len(m.terms.List) > 0 {
		s += m.terms.List[0].Definition + "\n\n"
	}

	return s
}

// Response Type
type Terms struct {
	List []struct {
		Definition  string    `json:"definition"`
		Permalink   string    `json:"permalink"`
		ThumbsUp    int       `json:"thumbs_up"`
		Author      string    `json:"author"`
		Word        string    `json:"word"`
		Defid       int       `json:"defid"`
		CurrentVote string    `json:"current_vote"`
		WrittenOn   time.Time `json:"written_on"`
		Example     string    `json:"example"`
		ThumbsDown  int       `json:"thumbs_down"`
	} `json:"list"`
}

// Cmd: talks to something outside of the event loop
func handleQuerySearch(q string) tea.Cmd {
	return func() tea.Msg {
		url := fmt.Sprintf("https://api.urbandictionary.com/v0/define?term=%s", url.QueryEscape(q))

		// cancel the http request is it istaking linger than 5 seconds
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return TermsResponseMessage{
				Err: err,
			}
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return TermsResponseMessage{
				Err: err,
			}
		}

		defer res.Body.Close()

		var wd Terms

		err = json.NewDecoder(res.Body).Decode(&wd)
		if err != nil {
			return TermsResponseMessage{
				Err: err,
			}
		}

		return TermsResponseMessage{
			Terms: wd,
		}
	}
}

// Msg
type TermsResponseMessage struct {
	Terms Terms
	Err   error
}
