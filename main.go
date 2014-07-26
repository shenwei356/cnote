package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/codegangsta/cli"
)

var (
	funcs  map[string]func(c *cli.Context)
	DBFILE string
	notedb *NoteDB
)

func init() {
	// DBFILE
	usr, err := user.Current()
	if err != nil {
		fmt.Println(err)
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

	funcs["dump"] = funDump
	funcs["wipe"] = funWipe
	funcs["restore"] = funRestore
	funcs["import"] = funImport

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
	if len(c.Args()) > 0 {
		fmt.Println("no arguments should be given.")
		return
	}

	for _, notename := range notedb.NotesList {

		// read note
		note, err := notedb.ReadNote(notename)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("note: %s\t(#. of items: %d, last update: %s).",
			notename, note.Sum, note.LastUpdate)
		if notedb.CurrentNote != nil &&
			notename == notedb.CurrentNote.NoteID {

			fmt.Printf(" (current note)")
		}
		fmt.Println()
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
		return
	}

	fmt.Printf("note \"%s\" created.\n", notename)
	fmt.Printf("current note: \"%s\" (last update: %s).\n",
		notename, notedb.CurrentNote.LastUpdate)
}

func funDel(c *cli.Context) {
	if len(c.Args()) == 0 {
		fmt.Println("note name needed.")
		return
	}
	notename := c.Args().First()

	note, err := notedb.ReadNote(notename)
	if err != nil {
		fmt.Println(err)
		return
	}

	reply, err := request_reply(
		fmt.Sprintf("==============================================================\n"+
			" Attention, it will delete all the %d items of note \"%s\".\n"+
			"==============================================================\n",
			note.Sum, notename)+
			" Type \"%s\" to continue:",
		"yes")
	if err != nil {
		fmt.Println(err)
		return
	}

	if reply == false {
		return
	}

	err = notedb.DeleteNote(notename)
	if err != nil {
		fmt.Println(err)
		return
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

	fmt.Printf("current note: \"%s\" (last update: %s).\n",
		notename, notedb.CurrentNote.LastUpdate)
}

func funAdd(c *cli.Context) {
	if len(c.Args()) != 2 {
		fmt.Println("tag and content needed.")
		return
	}

	item, err := notedb.AddNoteItem(c.Args()[0], c.Args()[1])
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(item)
}

func funRm(c *cli.Context) {
	if len(c.Args()) == 0 {
		fmt.Println("item ID needed.")
		return
	}

	for _, itemid := range c.Args() {

		itemid, err := strconv.Atoi(itemid)
		if err != nil {
			fmt.Println("item ID should be positive integer.")
			continue
		}

		// read item and print it, in case of misdeleteing
		item, err := notedb.ReadNoteItem(notedb.CurrentNote, itemid)
		if err != nil {
			fmt.Println(err)
			return
		}

		err = notedb.RemoveNoteItem(notedb.CurrentNote, itemid)
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Println(item)
	}
}

func funTag(c *cli.Context) {
	// list all tags
	note, err := notedb.GetCurrentNote()
	if err != nil {
		fmt.Println(err)
		return
	}

	if len(c.Args()) == 0 {
		tagstats := make([]TagStat, 0)
		for tag, taginfo := range note.Tags {
			tagstats = append(tagstats, TagStat{tag, len(taginfo)})
		}
		sort.Sort(SortTagsByAmount(tagstats))
		for _, tagstat := range tagstats {
			fmt.Printf("tag: %s\t(#. of items: %d).\n", tagstat.Tag, tagstat.Amount)
		}
		return
	}

	items, err := notedb.ItemByTag(c.Args())
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, item := range items {
		fmt.Println(item)
	}
}

func funSearch(c *cli.Context) {
	if len(c.Args()) == 0 {
		fmt.Println("search keyword needed.")
		return
	}

	items, err := notedb.ItemByRegexp(c.Args())
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, item := range items {
		fmt.Println(item)
	}
}

func funDump(c *cli.Context) {
	if len(c.Args()) > 0 {
		fmt.Println("no arguments should be given.")
		return
	}

	err := notedb.Dump()
	if err != nil {
		fmt.Println(err)
		return
	}
}

func funWipe(c *cli.Context) {
	if len(c.Args()) > 0 {
		fmt.Println("no arguments should be given.")
		return
	}

	reply, err := request_reply(
		"========================================\n"+
			" Attention, it will clear all the data.\n"+
			"========================================\n"+
			" Type \"%s\" to continue:",
		"yes")
	if err != nil {
		fmt.Println(err)
		return
	}

	if reply == false {
		return
	}

	err = notedb.Wipe()
	if err != nil {
		fmt.Println(err)
		return
	}
}

func funRestore(c *cli.Context) {
	if len(c.Args()) != 1 {
		fmt.Println("dumpped filename needed.")
		return
	}

	reply, err := request_reply(
		"========================================\n"+
			" Attention, it will clear all the data.\n"+
			"========================================\n"+
			" Type \"%s\" to continue:",
		"yes")
	if err != nil {
		fmt.Println(err)
		return
	}

	if reply == false {
		return
	}

	err = notedb.Restore(c.Args().First())
	if err != nil {
		fmt.Println(err)
		return
	}
}

func funImport(c *cli.Context) {
	if len(c.Args()) != 3 {
		fmt.Println("three arguments needed: <notename in your cnote>" +
			" <notename in dumpped note> <dumpped filename>.")
		return
	}
	notename, othernotename, filename := c.Args()[0], c.Args()[1], c.Args()[2]
	n, err := notedb.Import(notename, othernotename, filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%d items imported into note \"%s\".\n", n, notename)
}

func main() {
	notedb = NewNoteDB(DBFILE)
	defer notedb.Close()

	app := cli.NewApp()
	app.Name = "cnote"
	app.Usage = "A platform independent command line note app. https://github.com/shenwei356/cnote"
	app.Version = "1.2 (2014-07-22)"
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
			Usage:  "Add a note item",
			Action: getFunc(funcs, "add"),
		},
		{
			Name:   "rm",
			Usage:  "Remove a note item",
			Action: getFunc(funcs, "rm"),
		},
		{
			Name:      "tag",
			ShortName: "t",
			Usage:     "List items by tags. List all tags if no arguments given",
			Action:    getFunc(funcs, "tag"),
		},
		{
			Name:      "search",
			ShortName: "s",
			Usage:     "Search items with regular expression",
			Action:    getFunc(funcs, "search"),
		},
		{
			Name:   "dump",
			Usage:  "Dump whole database, for backup or transfer",
			Action: getFunc(funcs, "dump"),
		},
		{
			Name:   "wipe",
			Usage:  "Attention! Wipe whole database",
			Action: getFunc(funcs, "wipe"),
		},
		{
			Name:   "restore",
			Usage:  "Wipe whole database, and restore from dumpped file",
			Action: getFunc(funcs, "restore"),
		},
		{
			Name:   "import",
			Usage:  "Import note items from dumpped data",
			Action: getFunc(funcs, "import"),
		},
	}

	app.Run(os.Args)
}
