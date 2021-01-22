package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func main() {
	initialModel := getInitialModel()
	if len(os.Args) == 1 {
		log.Fatal("you must add a git command after gsel")
	}
	gitCmd := os.Args[1:]
	initialModel.cmd = gitCmd

	if len(gitCmd) == 0 {
		log.Fatal("you must add a valid git command after gsel")
	}

	p := tea.NewProgram(initialModel)

	if err := p.Start(); err != nil {
		fmt.Printf("there's been an error: %v", err)
		os.Exit(1)
	}
}

type model struct {
	choices []commit
	cursor int
	maxCursor int
	confirming bool
	cmd []string
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if (m.confirming) {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "enter":
				executeCmd(m.cmd, m.choices[m.cursor].hash)
				return m, tea.Quit
			}
		}
	} else {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.choices) - 1 {
					m.cursor++
					if m.cursor > m.maxCursor {
						m.maxCursor = m.cursor
					}
				}
			case "enter":
				m.confirming = true
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	
	if m.confirming {
		out := "Selected commit:"
		commit := m.choices[m.cursor]
		commitLine := prettifyCommit(false, commit) 

		out += commitLine + "\n"
		out += "Command: git " + strings.Join(m.cmd, " ") + " " + commit.hash

		return out
	}

	out := ""
	
	maxChoices := len(m.choices)
	// once we get 5 from the end, start displaying more commits
	// kinda like a pager

	var lastDisplayed int

	if maxChoices > 15 {
		lastDisplayed = 15
	} else {
		lastDisplayed = maxChoices - 1
	}

	if m.maxCursor > 10 && maxChoices > 15 {
		newLastDisplayed := lastDisplayed + m.maxCursor - 10
		if newLastDisplayed < maxChoices {
			lastDisplayed = newLastDisplayed
		} else {
			newLastDisplayed = maxChoices - 1
		}
	}

	for i, choice := range m.choices[0:lastDisplayed] {
		selected := m.cursor == i
		commitLine := prettifyCommit(selected, choice)
		out += commitLine
	}
	
	return out
}

func executeCmd(gitCmd []string, hash string) {
	args := append(gitCmd, hash)
	cmd := exec.Command("git", args...)
	fmt.Println(cmd)
	err := cmd.Run()
	checkErr(err)
}

func prettifyCommit(selected bool, c commit) string {
	cursor := " "

	red := color.New(color.FgRed)
	green := color.New(color.FgGreen)
	blue := color.New(color.FgBlue)

	if selected {
		cursor = ">"
		red.Add(color.Bold)
		green.Add(color.Bold)
		blue.Add(color.Bold)
	}

	strArr := []interface{}{
		cursor,
		red.Sprint(c.hash[0:8]),
		c.shortMessage,
		green.Sprint("(" + c.commitDate + ")"),
		blue.Sprint("<" + c.authorName + ">"),
	}

	commitLine := fmt.Sprintf("%s %s - %s %s %s\n", strArr...)
	return commitLine
}

func getInitialModel() (model) {
	commits := getCommits()

	return model{
		choices: commits,
		cursor: 0,
		confirming: false,
	}
}

func getCommits() []commit {
	directory, err := os.Getwd()
	checkErr(err)

	r, err := git.PlainOpen(directory)
	checkErr(err)

	head, err := r.Head()
	checkErr(err)

	logOptions := git.LogOptions{
		From: head.Hash(),
	}

	cIter, err := r.Log(&logOptions)
	checkErr(err)

	commits := []commit{}

	err = cIter.ForEach(func(c *object.Commit) error {
		newCommit := commit{
			authorName: c.Author.Name,
			commitDate: c.Author.When.Format("Jan 2, 03:04 PM"),
			hash: c.Hash.String(),
			shortMessage: strings.Split(c.Message, "\n")[0],
		}

		commits = append(commits, newCommit)
		return nil
	})
	
	checkErr(err)

	return commits
}


type commit struct {
	hash string
	shortMessage string
	commitDate string
	authorName string
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
