package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/fatih/color"
)

func main() {
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

	err = cIter.ForEach(func(c *object.Commit) error {
		shortHash := c.Hash.String()[0:8]
		shortMessage := strings.Split(c.Message, "\n")[0]
		commitDate := "(" + c.Author.When.Format("Jan 2, 03:04 PM") + ")"
		authorName := "<" + c.Author.Name + ">"

		strArr := []interface{}{
			color.RedString(shortHash),
			shortMessage,
			color.GreenString(commitDate),
			color.BlueString(authorName),
		}

		commitLine := fmt.Sprintf("%s - %s %s %s", strArr...)

		fmt.Println(commitLine)
		return nil
	})

	checkErr(err)
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
