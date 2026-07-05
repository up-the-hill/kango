package parser

import (
	"errors"
	"os"
	"strings"
)

type File struct {
	Frontmatter []string
	Main        []Column
	Config      []string
}

type Column struct {
	Title    string
	complete bool
	hidden   bool
	Cards    []Card
}

type Card struct {
	Done  bool
	Title string
}

func ParseFile(fPath string) (File, error) {
	f, err := os.ReadFile(fPath)
	if err != nil {
		// file doesn't exist
		return File{}, err
	}
	frontmatterCount := 0
	frontmatterEnd, mainEnd := -1, -1
	fileString := string(f)
	fileLines := strings.Split(fileString, "\n")
	for i, line := range fileLines {
		if line == "---" {
			frontmatterCount += 1
			if frontmatterCount == 2 {
				frontmatterEnd = i
			}
		}
		if line == "%% kanban:settings" {
			mainEnd = i
		}
	}
	if frontmatterEnd == -1 {
		return File{}, errors.New("file missing frontmatter")
	}
	if mainEnd == -1 {
		return File{}, errors.New("file missing config")
	}
	res := File{
		Frontmatter: fileLines[1:frontmatterEnd],
		Main:        parseColumns(fileLines[frontmatterEnd+1 : mainEnd]),
		Config:      fileLines[mainEnd:],
	}
	return res, nil
}

func parseColumns(lines []string) []Column {
	prev := -1
	res := []Column{}
	for i, line := range lines {
		// Check if line starts with hashtag
		if strings.HasPrefix(line, "#") {
			if prev == -1 {
				prev = i
				continue
			}
			res = append(res, parseColumn(lines[prev:i]))
			prev = i
		}
	}
	return res
}

func parseColumn(lines []string) Column {
	title := lines[0][3:]
	hidden := false
	complete := false
	cards := []Card{}
	for _, line := range lines[1:] {
		if strings.HasPrefix(line, "-") {
			cards = append(cards, parseCard(line))
		} else if line == "**Complete**" {
			complete = true
		}
	}

	return Column{
		Title:    title,
		complete: complete,
		hidden:   hidden,
		Cards:    cards,
	}
}

func parseCard(line string) Card {
	done := strings.HasPrefix(line, "- [x]")
	title := strings.TrimPrefix(line, "- [x] ")
	title = strings.TrimPrefix(title, "- [ ] ")
	return Card{
		Done:  done,
		Title: title,
	}
}
