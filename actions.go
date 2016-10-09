package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	b "github.com/barsanuphe/endive/book"
	e "github.com/barsanuphe/endive/endive"
	l "github.com/barsanuphe/endive/library"

	"github.com/codegangsta/cli"
)

func checkSortOrder(c *cli.Context) (orderDefined bool, sortBy string, lastIndex int) {
	if len(c.Args()) < 2 {
		return
	}
	sortBy = "default"
	for i, arg := range c.Args() {
		_, isIn := e.StringInSliceCaseInsensitive(arg, []string{"orderby", "sortby"})
		if isIn && i < c.NArg()-1 {
			// check args is valid
			if b.CheckValidSortOrder(c.Args()[i+1]) {
				orderDefined = true
				sortBy = strings.ToLower(c.Args()[i+1])
				lastIndex = i
				return
			}
		}
	}
	return
}

func checkLimits(c *cli.Context) (limitFirst, limitLast bool, number, lastIndex int) {
	if len(c.Args()) < 2 {
		return
	}
	for i, arg := range c.Args() {
		if i >= c.NArg()-1 {
			break
		}
		if strings.ToLower(arg) == "first" {
			limitFirst = true
		}
		if strings.ToLower(arg) == "last" {
			limitLast = true
		}
		if limitFirst || limitLast {
			// check c.Args()[i+1] is int
			nbr, err := strconv.Atoi(c.Args()[i+1])
			if err == nil {
				number = nbr
				lastIndex = i
				return
			}
		}
	}
	return
}

func checkArgsWithID(l l.Library, args []string) (book *b.Book, other []string, err error) {
	if len(args) < 1 {
		err = errors.New("Not enough arguments")
		return
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return
	}
	// get book from ID
	bk, err := l.Collection.FindByID(id)
	if err != nil {
		return nil, args[1:], err
	}
	book = bk.(*b.Book)
	other = args[1:]
	return
}

func editMetadata(c *cli.Context, endive *Endive) {
	book, args, err := checkArgsWithID(endive.Library, c.Args())
	if err != nil {
		endive.UI.Error("Error parsing arguments: " + err.Error())
		return
	}
	endive.UI.Title("Editing metadata for " + book.ShortString() + "\n")
	if err := book.EditField(args...); err != nil {
		endive.UI.Errorf("Error editing metadata for book ID#%d\n", book.ID())
	}
	_, _, err = book.Refresh()
	if err != nil {
		endive.UI.Errorf("Error refreshing book ID#%d\n", book.ID())
		return
	}
	showInfo(c, endive)
}

func refreshMetadata(c *cli.Context, endive *Endive) {
	book, _, err := checkArgsWithID(endive.Library, c.Args())
	if err != nil {
		endive.UI.Error("Error parsing arguments: " + err.Error())
		return
	}
	// is field specified?
	// TODO take list of fields as arguments?
	if len(c.Args()) == 2 {
		field := c.Args()[1]
		// check if valid field name
		_, isIn := e.StringInSlice(strings.ToLower(field), b.MetadataFieldNames)
		if !isIn {
			endive.UI.Error("Invalid metadata field " + field)
			return
		}
		// ask for confirmation
		if endive.UI.Accept("Confirm refreshing metadata field " + field + " for " + book.ShortString()) {
			err := book.ForceMetadataFieldRefresh(field)
			if err != nil {
				endive.UI.Errorf("Error reinitializing metadata field "+field+" for book ID#%d", book.ID)
				endive.UI.Error(err.Error())
			}
		}
	} else if len(c.Args()) == 1 {
		// ask for confirmation
		if endive.UI.Accept("Confirm refreshing metadata for " + book.ShortString()) {
			err := book.ForceMetadataRefresh()
			if err != nil {
				endive.UI.Errorf("Error reinitializing metadata for book ID#%d\n", book.ID())
			}
		}
	}
	_, _, err = book.Refresh()
	if err != nil {
		endive.UI.Errorf("Error refreshing book ID#%d\n", book.ID())
		return
	}
	showInfo(c, endive)
}

func setProgress(c *cli.Context, endive *Endive) {
	book, args, err := checkArgsWithID(endive.Library, c.Args())
	if err != nil {
		endive.UI.Error("Error parsing arguments: " + err.Error())
		return
	}
	if len(args) == 0 {
		// TODO or interactive mode?
		endive.UI.Error("Missing arguments. See help.")
		return
	}
	if len(args) >= 1 {
		// setting progress
		if err := book.SetProgress(args[0]); err != nil {
			endive.UI.Error("Progress must be among: unread/shortlisted/reading/read")
			return
		}
		// if first time set as read, set date too.
		if book.Progress == "read" && book.ReadDate == "" {
			book.SetReadDateToday()
		}
	}
	if len(args) >= 2 {
		// check rating format
		rating, e := strconv.ParseFloat(args[1], 32)
		if e != nil || rating < 0 || rating > 5 {
			endive.UI.Error("Rating must be a number between 0 and 5.")
			return
		}
		book.Rating = args[1]
	}
	if len(args) == 3 {
		book.Review = args[2]
	}
	showInfo(c, endive)
}

func showInfo(c *cli.Context, endive *Endive) {
	if c.NArg() == 0 {
		fmt.Println(endive.Library.ShowInfo())
	} else {
		book, _, err := checkArgsWithID(endive.Library, c.Args())
		if err != nil {
			endive.UI.Error("Error parsing arguments: " + err.Error())
			return
		}
		fmt.Println(book.ShowInfo())
	}
}

func listTags(c *cli.Context, endive *Endive) (err error) {
	book, _, err := checkArgsWithID(endive.Library, c.Args())
	if err != nil {
		// list all tags
		tags := endive.Library.Collection.Tags()
		endive.UI.Display(e.TabulateMap(tags, "Tag", "# of Books"))
	} else {
		// if ID, list tags of ID
		var rows [][]string
		rows = append(rows, []string{book.ShortString(), book.Metadata.Tags.String()})
		endive.UI.Display(e.TabulateRows(rows, "Book", "Tags"))
	}
	return
}

func listSeries(c *cli.Context, endive *Endive) {
	book, _, err := checkArgsWithID(endive.Library, c.Args())
	if err != nil {
		// list all series
		series := endive.Library.Collection.Series()
		endive.UI.Display(e.TabulateMap(series, "Series", "# of Books"))
	} else {
		// if ID, list series of ID
		var rows [][]string
		rows = append(rows, []string{book.ShortString(), book.Metadata.Series.String()})
		endive.UI.Display(e.TabulateRows(rows, "Book", "Series"))
	}
	return
}

func search(c *cli.Context, endive *Endive) {
	if c.NArg() == 0 {
		fmt.Println("No query found!")
	} else {
		order, sortBy, lastIndex1 := checkSortOrder(c)
		limitFirst, limitLast, number, lastIndex2 := checkLimits(c)
		queryParts := c.Args()
		if order || limitFirst || limitLast {
			// finding index of last argument part of search
			lastIndex := c.NArg()
			if order && lastIndex1 < lastIndex {
				lastIndex = lastIndex1
			}
			if (limitFirst || limitLast) && lastIndex2 < lastIndex {
				lastIndex = lastIndex2
			}
			// discard everything after "sortby"
			queryParts = queryParts[:lastIndex]
		}
		query := strings.Join(queryParts, " ")
		endive.UI.Debug("Searching for '" + query + "'...")
		var results e.Collection
		results = &b.Books{}
		hits, err := endive.Library.SearchAndPrint(query, sortBy, limitFirst, limitLast, number, results)
		if err != nil {
			endive.UI.Error(err.Error())
			panic(err)
		}
		endive.UI.Display(hits)
	}
}

func listImportableEpubs(endive *Endive, c *cli.Context, isRetail bool) {
	var candidates epubCandidates
	var err error
	var txt string

	if isRetail {
		candidates, err = endive.analyzeSources(endive.Config.RetailSource, isRetail)
		txt = fmt.Sprintf("Found %d retail epubs to import: ", len(candidates))
	} else {
		candidates, err = endive.analyzeSources(endive.Config.NonRetailSource, isRetail)
		txt = fmt.Sprintf("Found %d non-retail epubs to import: ", len(candidates))
	}
	if err != nil {
		endive.UI.Error(err.Error())
		return
	}
	if len(candidates) != 0 {
		endive.UI.SubPart(txt)
		for _, cd := range candidates {
			fmt.Println(" - " + cd.filename)
		}
	} else {
		endive.UI.SubPart("Nothing to import.")
	}
}

func importEpubs(endive *Endive, c *cli.Context, isRetail bool) {
	if len(c.Args()) >= 1 {
		// import valid paths
		if err := endive.ImportSpecific(isRetail, c.Args()...); err != nil {
			endive.UI.Error(err.Error())
			return
		}
	} else {
		var err error
		if isRetail {
			if len(endive.Config.RetailSource) == 0 {
				endive.UI.Error("No retail source found in configuration file!")
				return
			}
			err = endive.ImportRetail()
		} else {
			if len(endive.Config.NonRetailSource) == 0 {
				endive.UI.Error("No non-retail source found in configuration file!")
				return
			}
			err = endive.ImportNonRetail()
		}
		if err != nil {
			endive.UI.Error(err.Error())
		}
	}
}

func exportFilter(c *cli.Context, endive *Endive) {
	endive.UI.Title("Exporting selection to E-Reader...")
	query := strings.Join(c.Args(), " ")
	var books e.Collection
	books = &b.Books{}
	if err := endive.Library.Search(query, "default", false, false, 0, books); err != nil {
		endive.UI.Error("Error filtering books for export to e-reader")
		return
	}
	if err := endive.Library.ExportToEReader(books); err != nil {
		endive.UI.Errorf("Error exporting books to e-reader: %s", err.Error())
	}
}

func exportAll(endive *Endive) {
	endive.UI.Title("Exporting everything to E-Reader...")
	err := endive.Library.ExportToEReader(endive.Library.Collection)
	if err != nil {
		endive.UI.Errorf("Error exporting books to e-reader: %s", err.Error())
	}
}

func displayBooks(c *cli.Context, ui e.UserInterface, books e.Collection) {
	if sort, orderBy, _ := checkSortOrder(c); sort {
		books.Sort(orderBy)
	}
	limitFirst, limitLast, number, _ := checkLimits(c)
	if limitFirst {
		books = books.First(number)
	}
	if limitLast {
		books = books.Last(number)
	}
	ui.Display(books.Table())
}
