package endive

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/barsanuphe/gotabulate"
	i "github.com/barsanuphe/helpers/ui"
	"github.com/moraes/isbn"
)

// TabulateRows of map[string]int.
func TabulateRows(rows [][]string, headers ...string) (table string) {
	if len(rows) == 0 {
		return
	}
	t := gotabulate.Create(rows)
	t.SetHeaders(headers)
	t.SetEmptyString("N/A")
	t.SetAlign("left")
	t.SetAutoSize(true)
	return t.Render("border")
}

// TabulateMap of map[string]int.
func TabulateMap(input map[string]int, firstHeader string, secondHeader string) (table string) {
	if len(input) == 0 {
		return
	}
	// building first column list for sorting
	var keys []string
	for key := range input {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	// building table
	var rows [][]string
	for _, key := range keys {
		rows = append(rows, []string{key, strconv.Itoa(input[key])})
	}
	return TabulateRows(rows, firstHeader, secondHeader)
}

// CleanISBN from a string
func CleanISBN(full string) (isbn13 string, err error) {
	// cleanup string, only keep numbers
	re := regexp.MustCompile("[0-9]+X?")
	candidate := strings.Join(re.FindAllString(strings.ToUpper(full), -1), "")

	// if start of isbn detected, try to salvage the situation
	if len(candidate) > 13 && strings.HasPrefix(candidate, "978") {
		candidate = candidate[:13]
	}

	// validate and convert to ISBN13 if necessary
	if isbn.Validate(candidate) {
		if len(candidate) == 10 {
			isbn13, err = isbn.To13(candidate)
			if err != nil {
				isbn13 = ""
			}
		}
		if len(candidate) == 13 {
			isbn13 = candidate
		}
	} else {
		err = errors.New("ISBN-13 not found")
	}
	return
}

// AskForISBN when not found in epub
func AskForISBN(ui i.UserInterface) (string, error) {
	if ui.Accept("Do you want to enter an ISBN manually") {
		validChoice := false
		errs := 0
		for !validChoice {
			fmt.Print("Enter ISBN: ")
			choice, scanErr := ui.GetInput()
			if scanErr != nil {
				return "", scanErr
			}
			// check valid ISBN
			isbnCandidate, err := CleanISBN(choice)
			if err != nil {
				errs++
				ui.Warning("Warning: Invalid value.")
			} else {
				confirmed := ui.Accept("Confirm: " + choice)
				if confirmed {
					validChoice = true
					return isbnCandidate, nil
				}
				errs++
				fmt.Println("Manual entry not confirmed, trying again.")
			}
			if errs > 5 {
				ui.Warning("Too many errors, continuing without ISBN.")
				break
			}
		}
	}
	return "", errors.New("ISBN not set")
}
