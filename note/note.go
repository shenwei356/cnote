package note

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/jinzhu/now"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

var (
	NOTE_PREFIX = "note_"
	ITEM_PREFIX = "item_"
)

func trim_prefix(p, s string) string {
	return regexp.MustCompile("^"+p).ReplaceAllString(s, "")
}

type Config struct {
	CurrentNoteName string `json:"current_note_name"`
}

type Note struct {
	NoteID     string                     `json:"noteid"`
	Sum        int                        `json:"sum"`
	LastUpdate string                     `json:"last_update"`
	LastId     int                        `json:"last_id"`
	Tags       map[string]map[string]bool `json:"tags"`

	Items map[int]*Item `json:"-"`
}

type Item struct {
	ItemID  string   `json:"itemid"`
	Tags    []string `json:"tags"`
	Content string   `json:"content"`
}

type NoteDB struct {
	Config      *Config
	NotesList   []string
	CurrentNote *Note

	db     *leveldb.DB
	dbfile string
}

func NewNoteDB(dbfile string) *NoteDB {
	notedb := new(NoteDB)
	notedb.dbfile = dbfile
	notedb.ConnectDB()
	notedb.ReadConfig()

	err := notedb.UseNote(notedb.Config.CurrentNoteName)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	list, err := notedb.GetNotesList()
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	notedb.NotesList = list

	/*	if len(list) == 0 {
		fmt.Println("no notes in database, please create one.")
	}*/

	return notedb
}

//////////////////////////////////////////////////////
func (notedb *NoteDB) ConnectDB() {
	db, err := leveldb.OpenFile(notedb.dbfile, nil)
	if err != nil {
		fmt.Sprintf("fail to open leveldb file: %s. %s", notedb.dbfile, err)
		os.Exit(0)
	}
	notedb.db = db
}

func (notedb *NoteDB) Close() {
	notedb.SaveConfig()
	notedb.db.Close()
}

//////////////////////////////////////////////////////
func (notedb *NoteDB) NewNote(notename string) error {
	var note = &Note{}
	key := NOTE_PREFIX + notename
	err := notedb.ReadStruct(key, note)
	if err == nil {
		return errors.New(fmt.Sprintf("note \"%s\" already exist.", notename))
	}

	note = &Note{
		NoteID:     notename,
		Sum:        0,
		LastUpdate: now.BeginningOfMinute().String(),
		LastId:     0,
		Tags:       map[string]map[string]bool{},
	}

	notedb.NotesList = append(notedb.NotesList, notename)
	err = notedb.SaveStruct(key, note)
	if err != nil {
		return errors.New(fmt.Sprintf("fail to save %s. %v", key, err))
	}

	err = notedb.UseNote(notename)
	if err != nil {
		return err
	}
	return nil
}

func (notedb *NoteDB) DelNote(notename string) error {
	var note = &Note{}
	key := NOTE_PREFIX + notename
	err := notedb.ReadStruct(key, note)
	if err != nil {
		return errors.New(fmt.Sprintf("note \"%s\" not exist in %v.", notename, notedb.NotesList))
	}

	// remove items

	for tag, _ := range note.Tags {
		for itemid, _ := range note.Tags[tag] {
			itemid, err := strconv.Atoi(itemid)
			if err != nil {
				fmt.Println(err)
				continue
			}

			notedb.RemoveNoteItem(note, itemid)
		}
	}

	err = notedb.DeleteStruct(key)
	if err != nil {
		return err
	}

	// update list
	list := make([]string, 0)
	for _, n := range notedb.NotesList {
		if n != notename {
			list = append(list, n)
		}
	}
	notedb.NotesList = list

	// update config
	notedb.Config.CurrentNoteName = ""
	notedb.CurrentNote = nil
	notedb.SaveConfig()
	return nil
}

func (notedb *NoteDB) UseNote(notename string) error {
	var note = &Note{}

	if notename == "" && len(notedb.NotesList) == 0 {
		notedb.Config.CurrentNoteName = notename
		notedb.CurrentNote = nil
		return nil
	}

	err := notedb.ReadStruct(NOTE_PREFIX+notename, note)
	if err != nil {
		return errors.New(
			fmt.Sprintf("note \"%s\" not exist in %v.",
				notename, notedb.NotesList))
	}

	notedb.Config.CurrentNoteName = notename
	notedb.CurrentNote = note

	return nil
}

func (notedb *NoteDB) GetNotesList() ([]string, error) {
	list := make([]string, 0)
	iter := notedb.db.NewIterator(&util.Range{Start: []byte(NOTE_PREFIX)}, nil)
	for iter.Next() {
		key := iter.Key()
		list = append(list, trim_prefix(NOTE_PREFIX, string(key)))
	}
	iter.Release()
	err := iter.Error()
	return list, err
}

func (notedb *NoteDB) AddNoteItem(tagstring, content string) error {
	if notedb.CurrentNote == nil {
		return errors.New("no note choosed.")
	}

	tags := strings.Split(tagstring, ",")

	notedb.CurrentNote.LastId++

	item := &Item{
		ItemID:  fmt.Sprintf("%d", notedb.CurrentNote.LastId),
		Tags:    tags,
		Content: content,
	}

	// save item
	key := fmt.Sprintf("%s%s_%09d", ITEM_PREFIX,
		notedb.CurrentNote.NoteID, notedb.CurrentNote.LastId)
	err := notedb.SaveStruct(key, item)
	if err != nil {
		return errors.New(fmt.Sprintf("fail to save %s. %v", key, err))
	}

	// update note
	notedb.CurrentNote.Sum++
	for _, tag := range tags {
		if _, ok := notedb.CurrentNote.Tags[tag]; !ok {
			notedb.CurrentNote.Tags[tag] = make(map[string]bool, 0)
		}
		notedb.CurrentNote.Tags[tag][item.ItemID] = true
	}
	notedb.CurrentNote.LastUpdate = now.BeginningOfMinute().String()
	key = NOTE_PREFIX + notedb.CurrentNote.NoteID
	err = notedb.SaveStruct(key, notedb.CurrentNote)
	if err != nil {
		return errors.New(fmt.Sprintf("fail to save %s. %v", key, err))
	}

	return nil
}

func (notedb *NoteDB) GetNoteItem(note *Note, itemid int) (*Item, error) {
	if note == nil {
		return nil, errors.New("no note choosed.")
	}

	var item = &Item{}
	key := fmt.Sprintf("%s%s_%09d", ITEM_PREFIX,
		note.NoteID, itemid)
	err := notedb.ReadStruct(key, item)
	if err != nil {
		return nil, errors.New(
			fmt.Sprintf("item \"%d\" not exist in note \"%s\".",
				itemid, note.NoteID))
	}

	return item, nil
}

func (notedb *NoteDB) RemoveNoteItem(note *Note, itemid int) error {
	item, err := notedb.GetNoteItem(note, itemid)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%s%s_%09d", ITEM_PREFIX,
		note.NoteID, itemid)

	err = notedb.DeleteStruct(key)
	if err != nil {
		return err
	}

	// update note
	note.Sum--
	for _, tag := range item.Tags {
		delete(note.Tags[tag], item.ItemID)
		if len(note.Tags[tag]) == 0 {
			delete(note.Tags, tag)
		}
	}
	note.LastUpdate = now.BeginningOfMinute().String()
	key = NOTE_PREFIX + note.NoteID
	err = notedb.SaveStruct(key, note)
	if err != nil {
		return errors.New(
			fmt.Sprintf("fail to save %s. %v", key, err))
	}

	return nil
}

func (notedb *NoteDB) Search(queries []string) error {
	if notedb.CurrentNote == nil {
		return errors.New("no note choosed.")
	}

	note := notedb.CurrentNote
	if note.Items == nil {
		note.Items = make(map[int]*Item, 0)
	}
	for tag, _ := range note.Tags {
		for itemid, _ := range note.Tags[tag] {
			itemid, err := strconv.Atoi(itemid)
			if err != nil {
				fmt.Println(err)
				continue
			}

			if _, ok := note.Items[itemid]; ok { // loaded
				continue
			}
			item, err := notedb.GetNoteItem(note, itemid)
			if err != nil {
				return err
			}
			note.Items[itemid] = item
		}
	}

	for _, query := range queries {
		re := regexp.MustCompile(query)
		for _, item := range note.Items {
			if re.MatchString(item.Content) {
				print_item(item)
			}
		}
	}

	return nil
}

func (notedb *NoteDB) ListByTag(tags []string) error {
	if notedb.CurrentNote == nil {
		return errors.New("no note choosed.")
	}

	note := notedb.CurrentNote

	// list all tags
	if len(tags) == 0 {
		for tag, _ := range note.Tags {
			fmt.Printf("%s\n", tag)
		}
		return nil
	}

	for _, tag := range tags {
		if _, ok := note.Tags[tag]; !ok {
			fmt.Printf("tag \"%s\" not exist in note \"%s\".\n", tag, note.NoteID)
			continue
		}

		itemids := make([]int, 0)
		for itemid, _ := range note.Tags[tag] {
			itemid, err := strconv.Atoi(itemid)
			if err != nil {
				fmt.Println(err)
				continue
			}

			itemids = append(itemids, itemid)
		}

		for _, itemid := range itemids {
			item, err := notedb.GetNoteItem(note, itemid)
			if err != nil {
				return err
			}
			print_item(item)
		}
	}

	return nil
}

func print_item(item *Item) {
	fmt.Printf("%s: %s\n", item.ItemID, item.Content)
}

//////////////////////////////////////////////////////////////////////

func (notedb *NoteDB) ReadConfig() {
	notedb.Config = &Config{CurrentNoteName: ""}
	err := notedb.ReadStruct("config", notedb.Config)
	if err == nil {
		return
	}
	notedb.Config = &Config{CurrentNoteName: ""}
}

func (notedb *NoteDB) SaveConfig() {
	err := notedb.SaveStruct("config", notedb.Config)
	if err != nil {
		fmt.Sprintf("fail to save config. %v", err)
		os.Exit(0)
	}
}

//////////////////////////////////////////////////////////////////////

func (notedb *NoteDB) ReadStruct(key string, str interface{}) error {
	data, err := notedb.db.Get([]byte(key), nil)
	if err != nil { // not exist
		return err
	}
	if err := json.Unmarshal(data, str); err != nil {
		return err
	}

	return nil
}

func (notedb *NoteDB) SaveStruct(key string, str interface{}) error {
	bytes, err := json.Marshal(str)
	if err != nil {
		return err
	}

	err = notedb.db.Put([]byte(key), bytes, nil)
	if err != nil {
		return err
	}

	// fmt.Printf("save %s: %s\n", key, string(bytes))
	return nil
}

func (notedb *NoteDB) DeleteStruct(key string) error {
	err := notedb.db.Delete([]byte(key), nil)
	if err != nil {
		return err
	}
	return nil
}

func (notedb *NoteDB) Dump() error {
	iter := notedb.db.NewIterator(nil, nil)
	for iter.Next() {
		key := iter.Key()
		value := iter.Value()
		fmt.Printf("\"%s\":%s,\n", key, value)
	}
	iter.Release()
	err := iter.Error()
	return err
}
