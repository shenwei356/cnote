package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"github.com/codegangsta/cli"
	"github.com/shenwei356/cnote/note"
)

var (
	funcs  map[string]func(c *cli.Context)
	DBFILE string
	notedb *note.NoteDB
)

func init() {
	// DBFILE
	usr, err := user.Current()
	if err != nil {
		os.Exit(1)
	}
	DBFILE = filepath.Join(usr.HomeDir, ".cnote")

	funcs = make(map[string]func(c *cli.Context))
	funcs["new"] = funNew
	funcs["del"] = funDel
	funcs["use"] = funUse
	funcs["list"] = funLs
	funcs["add"] = funAdd
	funcs["rm"] = funRm
	funcs["tag"] = funTag
	funcs["search"] = funSearch

}

func getFunc(funcs map[string]func(c *cli.Context), name string) func(c *cli.Context) {
	if f, ok := funcs[name]; ok {
		return f
	} else {
		return func(c *cli.Context) {
			fmt.Printf("command %s not implemented\n", name)
		}
	}
}

func funLs(c *cli.Context) {
	list := notedb.NotesList
	for _, notename := range list {
		fmt.Println(notename)
	}
}

func funNew(c *cli.Context) {
	if len(c.Args()) == 0 {
		fmt.Println("note name needed.")
		return
	}
	if len(c.Args()) > 1 {
		fmt.Println("only one note name allowed.")
		return
	}

	notename := c.Args().First()

	err := notedb.NewNote(notename)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("note \"%s\" created.\n", notename)
	fmt.Printf("current note: \"%s\".\n", notename)
}

func funDel(c *cli.Context) {
	if len(c.Args()) == 0 {
		fmt.Println("note name needed.")
		return
	}

	for _, notename := range c.Args() {
		err := notedb.DelNote(notename)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// fmt.Printf("note \"%s\" deleted.\n", notename)
	}
}

func funUse(c *cli.Context) {
	if len(c.Args()) == 0 {
		fmt.Println("note name needed.")
		return
	}
	if len(c.Args()) > 1 {
		fmt.Println("only one note name allowed.")
		return
	}

	notename := c.Args().First()
	err := notedb.UseNote(notename)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("current note: \"%s\".\n", notename)
}

func funAdd(c *cli.Context) {
	if len(c.Args()) != 2 {
		fmt.Println("tag and content needed.")
		return
	}

	err := notedb.AddNoteItem(c.Args()[0], c.Args()[1])
	if err != nil {
		fmt.Println(err)
	}
}

func funRm(c *cli.Context) {
	if len(c.Args()) == 0 {
		fmt.Println("item ID needed.")
		return
	}

	for _, itemid := range c.Args() {

		itemid, err := strconv.Atoi(itemid)
		if err != nil {
			fmt.Println("item ID should be integer.")
			continue
		}

		err = notedb.RemoveNoteItem(notedb.CurrentNote, itemid)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// fmt.Printf("note item \"%d\" deleted from note \"%s\".\n", itemid, notedb.CurrentNote.NoteID)
	}
}

func funTag(c *cli.Context) {
	err := notedb.ListByTag(c.Args())
	if err != nil {
		fmt.Println(err)
		return
	}
}

func funSearch(c *cli.Context) {
	if len(c.Args()) == 0 {
		fmt.Println("search keyword needed.")
		return
	}
	err := notedb.Search(c.Args())
	if err != nil {
		fmt.Println(err)
		return
	}
}

func main() {
	notedb = note.NewNoteDB(DBFILE)

	app := cli.NewApp()
	app.Name = "cnote"
	app.Usage = "A command line note app. https://github.com/shenwei356/cnote"
	app.Version = "1.0 (2014-07-18)"
	app.Author = "Wei Shen"
	app.Email = "shenwei356@gmail.com"

	app.Commands = []cli.Command{
		{
			Name:   "new",
			Usage:  "Create a new note",
			Action: getFunc(funcs, "new"),
		},
		{
			Name:   "del",
			Usage:  "Delete a note",
			Action: getFunc(funcs, "del"),
		},
		{
			Name:   "use",
			Usage:  "Select a note",
			Action: getFunc(funcs, "use"),
		},
		{
			Name:      "list",
			ShortName: "ls",
			Usage:     "List all notes",
			Action:    getFunc(funcs, "list"),
		},
		{
			Name:   "add",
			Usage:  "add a note item",
			Action: getFunc(funcs, "add"),
		},
		{
			Name:   "rm",
			Usage:  "Remove a note item",
			Action: getFunc(funcs, "rm"),
		},
		{
			Name:      "search",
			ShortName: "s",
			Usage:     "Search for",
			Action:    getFunc(funcs, "search"),
		},
		{
			Name:      "tag",
			ShortName: "t",
			Usage:     "List items by tags",
			Action:    getFunc(funcs, "tag"),
		},
		{
			Name:  "dump",
			Usage: "Dump whole database",
			Action: func(c *cli.Context) {
				notedb.Dump()
			},
		},
	}

	app.Run(os.Args)

	notedb.Close()
}
