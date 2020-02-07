package main

import (
	"fmt"
	"github.com/gramarr/radarr"
	tb "gopkg.in/tucnak/telebot.v2"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// HandleAddMovie func
func (e *Env) HandleListMovies(m *tb.Message) {
	e.CM.StartConversation(NewListMoviesConversation(e), m)
}

// NewAddMovieConversation func
func NewListMoviesConversation(e *Env) *ListMoviesConversation {
	return &ListMoviesConversation{env: e}
}

// AddMovieConversation struct
type ListMoviesConversation struct {
	currentStep            Handler
	movieQuery             string
	movieResults           []radarr.Movie
	folderResults          []radarr.Folder
	selectedMovie          *radarr.Movie
	selectedQualityProfile *radarr.Profile
	selectedFolder         *radarr.Folder
	env                    *Env
}

// Run func
func (c *ListMoviesConversation) Run(m *tb.Message) {
	c.currentStep = c.AskFolder(m)
}

// Name func
func (c *ListMoviesConversation) Name() string {
	return "listMovies"
}

// CurrentStep funcfunc
func (c *ListMoviesConversation) CurrentStep() Handler {
	return c.currentStep
}

func (c *ListMoviesConversation) AskFolder(m *tb.Message) Handler {

	folders, err := c.env.Radarr.GetFolders(true)
	c.folderResults = folders

	// GetFolders Service Failed
	if err != nil {
		SendError(c.env.Bot, m.Sender, "Failed to get folders.")
		c.env.CM.StopConversation(c)
		return nil
	}

	// No Results
	if len(folders) == 0 {
		SendError(c.env.Bot, m.Sender, "No destination folders found.")
		c.env.CM.StopConversation(c)
		return nil
	}

	// Found folders!

	// Send the results
	var msg []string
	msg = append(msg, "*Available folders:*")
	for _, folder := range folders {
		msg = append(msg, fmt.Sprintf("- %s", EscapeMarkdown(strings.Title(filepath.Base(folder.Path)))))
	}
	Send(c.env.Bot, m.Sender, strings.Join(msg, "\n"))

	// Send the custom reply keyboard
	var options []string
	options = append(options, "All")
	for _, folder := range folders {
		options = append(options, fmt.Sprintf("%s", strings.Title(filepath.Base(folder.Path))))
	}
	options = append(options, "/cancel")
	SendKeyboardList(c.env.Bot, m.Sender, "Which movies would you like to list?", options)

	return func(m *tb.Message) {
		// Set the selected folder
		for i, opt := range options {
			if m.Text == opt {
				if m.Text == "All" {
					c.selectedFolder = &radarr.Folder{
						Path:      "",
						FreeSpace: -1,
						ID:        -1,
					}
				} else {
					c.selectedFolder = &c.folderResults[i-1]
				}
				break
			}
		}

		if c.selectedFolder == nil {
			SendError(c.env.Bot, m.Sender, "Invalid selection.")
			c.currentStep = c.AskFolder(m)
			return
		}
		c.currentStep = c.AskMovie(m)
	}
}

// HandleAddMovie func
func (c *ListMoviesConversation) AskMovie(m *tb.Message) Handler {
	c.movieResults, _ = c.env.Radarr.GetMoviesFromFolder(*c.selectedFolder)

	sort.Slice(c.movieResults, func(i, j int) bool {
		return c.movieResults[i].Title < c.movieResults[j].Title
	})

	var fulfilled = []string{"*Available Movies:*"}
	var pending = []string{"*Pending Movies:*"}
	var options []string
	for _, movie := range c.movieResults {
		options = append(options, EscapeMarkdown(movie.Title))
		if movie.HasFile {
			fulfilled = append(fulfilled, fmt.Sprintf("- %s", EscapeMarkdown(movie.Title)))
		} else {
			pending = append(pending, fmt.Sprintf("- %s", EscapeMarkdown(movie.Title)))
		}
	}
	if len(fulfilled) > 1 {
		Send(c.env.Bot, m.Sender, strings.Join(fulfilled, "\n"))
	}
	if len(pending) > 1 {
		Send(c.env.Bot, m.Sender, strings.Join(pending, "\n"))
	}

	if len(options) > 0 {
		options = append(options, "/cancel")
		SendKeyboardList(c.env.Bot, m.Sender, "Select a movie for more details or send [/cancel]", options)
	} else {
		Send(c.env.Bot, m.Sender, "No movies found")
		c.currentStep = c.AskFolder(m)
	}

	return func(m *tb.Message) {

		// Set the selected movie
		for i, opt := range options {
			if m.Text == opt {
				c.selectedMovie = &c.movieResults[i]
				break
			}
		}

		// Not a valid movie selection
		if c.selectedMovie == nil {
			SendError(c.env.Bot, m.Sender, "Invalid selection.")
			return
		}

		m.Payload = strconv.Itoa(c.selectedMovie.ID)
		c.env.HandleDetails(m)
		c.env.CM.StopConversation(c)
	}

}
