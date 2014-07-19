package cnotedb

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
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

type Config struct {
	CurrentNoteName string `json:"current_note_name"`
}

type Note struct {
	NoteID     string `json:"noteid"`
	Sum        int    `json:"sum"`
	LastUpdate string `json:"last_update"`
	LastId     int    `json:"last_id"`
	// not that, the type of the key of interal map is string
	// because the leveldb only allow string as key.
	Tags map[string]map[string]bool `json:"tags"`

	Items map[int]*Item `json:"-"`
}

type Item struct {
	ItemID  string   `json:"itemid"`
	Tags    []string `json:"tags"`
	Content string   `json:"content"`
}

func (item *Item) String() string {
	return fmt.Sprintf("item: %s\t(tags: %v)\t%s",
		item.ItemID, item.Tags, item.Content)
}

/*
type SortItemsById []Item

func (items SortItemsById) Len() int {
	return len(items)
}
func (items SortItemsById) Swap(i, j int) {
	items[i], items[j] = items[j], items[i]
}

func (items SortItemsById) Less(i, j int) bool {
	return items[i].ItemID < items[j].ItemID
}
*/
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

	// check the config
	_, err := notedb.ReadNote(notedb.Config.CurrentNoteName)
	if err != nil {
		notedb.Config.CurrentNoteName = ""
	}

	err = notedb.UseNote(notedb.Config.CurrentNoteName)
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

func (notedb *NoteDB) GetCurrentNote() (*Note, error) {
	if notedb.CurrentNote == nil {
		return nil, errors.New(
			fmt.Sprintf("no note choosed from %v. Use \"cnote use notename\".",
				notedb.NotesList))
	}
	return notedb.CurrentNote, nil
}

func (notedb *NoteDB) ReadNote(notename string) (*Note, error) {
	var note = &Note{}
	key := NOTE_PREFIX + notename

	err := notedb.ReadStruct(key, note)
	if err != nil {
		return nil, errors.New(
			fmt.Sprintf("note \"%s\" not exist.", notename))
	}

	return note, nil
}

func (notedb *NoteDB) SaveNote(note *Note) error {
	key := NOTE_PREFIX + note.NoteID
	err := notedb.SaveStruct(key, note)
	if err != nil {
		return errors.New(fmt.Sprintf("fail to save %s. %v", key, err))
	}

	return nil
}

func (notedb *NoteDB) NewNote(notename string) error {

	// check whether note exists
	_, err := notedb.ReadNote(notename)
	if err == nil {
		return errors.New(
			fmt.Sprintf("note \"%s\" already exist.", notename))
	}

	note := &Note{
		NoteID:     notename,
		Sum:        0,
		LastUpdate: now.BeginningOfMinute().String(),
		LastId:     0,
		Tags:       map[string]map[string]bool{},
	}

	notedb.NotesList = append(notedb.NotesList, notename)

	// save note
	err = notedb.SaveNote(note)
	if err != nil {
		return err
	}

	err = notedb.UseNote(notename)
	if err != nil {
		return err
	}
	return nil
}

func (notedb *NoteDB) DeleteNote(notename string) error {

	// read note
	note, err := notedb.ReadNote(notename)
	if err != nil {
		return err
	}

	// first, remove all items of the note
	itemids := make(map[int]bool, 0)
	for tag, _ := range note.Tags {
		for itemid, _ := range note.Tags[tag] {

			itemid, err := strconv.Atoi(itemid)
			if err != nil {
				return err
			}

			itemids[itemid] = true
		}
	}
	for itemid, _ := range itemids {
		err := notedb.RemoveNoteItem(note, itemid)
		if err != nil {
			return err
		}
	}

	// second, delete the note
	key := NOTE_PREFIX + note.NoteID
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

	return nil
}

func (notedb *NoteDB) UseNote(notename string) error {

	// not note exists
	if notename == "" && len(notedb.NotesList) == 0 {
		notedb.Config.CurrentNoteName = notename
		notedb.CurrentNote = nil
		return nil
	}

	// read note
	note, err := notedb.ReadNote(notename)
	if err != nil {
		return err
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

func (notedb *NoteDB) AddNoteItem(tagstring, content string) (*Item, error) {
	note, err := notedb.GetCurrentNote()
	if err != nil {
		return nil, err
	}

	tags := make([]string, 0)
	re := regexp.MustCompile(`^\s*$`)
	for _, tag := range strings.Split(tagstring, ",") {
		// empty
		if re.MatchString(tag) {
			continue
		}

		tags = append(tags, tag)
	}

	notedb.CurrentNote.LastId++

	item := &Item{
		ItemID:  fmt.Sprintf("%d", notedb.CurrentNote.LastId),
		Tags:    tags,
		Content: content,
	}

	// save item
	key := fmt.Sprintf("%s%s_%09d", ITEM_PREFIX,
		notedb.CurrentNote.NoteID, notedb.CurrentNote.LastId)
	err = notedb.SaveStruct(key, item)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("fail to save %s. %v", key, err))
	}

	// update current note
	note.Sum++
	for _, tag := range tags {
		if _, ok := note.Tags[tag]; !ok {
			note.Tags[tag] = make(map[string]bool, 0)
		}

		note.Tags[tag][item.ItemID] = true
	}
	note.LastUpdate = now.BeginningOfMinute().String()

	// save note
	err = notedb.SaveNote(note)
	if err != nil {
		return nil, err
	}

	return item, nil
}

func (notedb *NoteDB) ReadNoteItem(note *Note, itemid int) (*Item, error) {
	if note == nil {
		return nil, errors.New(
			fmt.Sprintf("no note choosed from %v. Use \"cnote use notename\".",
				notedb.NotesList))
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
	item, err := notedb.ReadNoteItem(note, itemid)
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

	// save note
	err = notedb.SaveNote(note)
	if err != nil {
		return err
	}

	return nil
}

func (notedb *NoteDB) ItemByRegexp(queries []string) ([]*Item, error) {
	note, err := notedb.GetCurrentNote()
	if err != nil {
		return nil, err
	}

	// read all items
	if note.Items == nil {
		note.Items = make(map[int]*Item, 0)
	}
	for tag, _ := range note.Tags {
		for itemid, _ := range note.Tags[tag] {
			itemid, err := strconv.Atoi(itemid)
			if err != nil {
				return nil, err
			}

			if _, ok := note.Items[itemid]; ok { // loaded
				continue
			}

			item, err := notedb.ReadNoteItem(note, itemid)
			if err != nil {
				return nil, err
			}

			note.Items[itemid] = item
		}
	}

	// query by regexp
	items := make([]*Item, 0)
	for _, query := range queries {
		re := regexp.MustCompile(query)

		itemids := make([]int, 0)
		for itemid, _ := range note.Items {
			itemids = append(itemids, itemid)
		}

		sort.Ints(itemids)

		for _, itemid := range itemids {
			item := note.Items[itemid]
			if re.MatchString(item.Content) {
				items = append(items, item)
			}
		}
	}

	return items, nil
}

func (notedb *NoteDB) ItemByTag(tags []string) ([]*Item, error) {
	note, err := notedb.GetCurrentNote()
	if err != nil {
		return nil, err
	}

	items := make([]*Item, 0)

	for _, tag := range tags {
		if _, ok := note.Tags[tag]; !ok {
			fmt.Printf("tag \"%s\" not exist in note \"%s\".\n", tag, note.NoteID)
			continue
		}

		itemids := make([]int, 0)
		for itemid, _ := range note.Tags[tag] {
			itemid, err := strconv.Atoi(itemid)
			if err != nil {
				return nil, err
			}

			itemids = append(itemids, itemid)
		}

		sort.Ints(itemids)

		for _, itemid := range itemids {
			item, err := notedb.ReadNoteItem(note, itemid)
			if err != nil {
				return nil, err
			}

			items = append(items, item)
		}
	}

	return items, nil
}

//////////////////////////////////////////////////////////////////////

func (notedb *NoteDB) ReadConfig() {
	notedb.Config = &Config{}

	err := notedb.ReadStruct("config", notedb.Config)
	if err != nil { // no config
		notedb.Config = &Config{CurrentNoteName: ""}
		return
	}
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
		fmt.Printf("%s\t%s\n", key, value)
	}
	iter.Release()
	err := iter.Error()
	return err
}

func (notedb *NoteDB) Wipe() error {
	for _, notename := range notedb.NotesList {
		err := notedb.DeleteNote(notename)
		if err != nil {
			return err
		}
	}
	return nil
}

func (notedb *NoteDB) Restore(filename string) error {
	// wipe all the database
	err := notedb.Wipe()
	if err != nil {
		return err
	}

	fh, err := os.Open(filename)
	if err != nil {
		return errors.New("fail to open file: " + filename)
	}

	batch := new(leveldb.Batch)
	reader := bufio.NewReader(fh)
	re1 := regexp.MustCompile(`[\r\n]`)
	re2 := regexp.MustCompile(`^\s+|\s+$`)
	re := regexp.MustCompile(`([^\t]+)\t([^\t]+)`)
	for {
		str, err := reader.ReadString('\n')

		str = re1.ReplaceAllString(str, "")
		str = re2.ReplaceAllString(str, "")
		if err == io.EOF {

			if re.MatchString(str) {
				data := re.FindSubmatch([]byte(str))
				batch.Put(data[1], data[2])
			}

			break
		}

		if re.MatchString(str) {
			data := re.FindSubmatch([]byte(str))
			batch.Put(data[1], data[2])
		}
	}

	err = notedb.db.Write(batch, nil)
	if err != nil {
		return nil
	}

	return nil
}

func (notedb *NoteDB) Import(notename, othernotename, filename string) (int, error) {
	err := notedb.UseNote(notename)
	if err != nil {
		return 0, err
	}

	fh, err := os.Open(filename)
	if err != nil {
		return 0, errors.New("fail to open file: " + filename)
	}

	batch := new(leveldb.Batch)
	reader := bufio.NewReader(fh)
	re1 := regexp.MustCompile(`[\r\n]`)
	re2 := regexp.MustCompile(`^\s+|\s+$`)
	re := regexp.MustCompile(`([^\t]+)\t([^\t]+)`)

	n := 0

	for {
		str, err := reader.ReadString('\n')

		str = re1.ReplaceAllString(str, "")
		str = re2.ReplaceAllString(str, "")
		if err == io.EOF {

			if !re.MatchString(str) {
				break
			}

			data := re.FindSubmatch([]byte(str))
			key := string(data[1])
			value := data[2]

			if !strings.HasPrefix(key, ITEM_PREFIX) {
				break
			}

			key = trim_prefix(ITEM_PREFIX, key)
			if !strings.HasPrefix(key, othernotename) {
				break
			}

			item := &Item{}
			if err := json.Unmarshal(value, item); err != nil {
				return 0, err
			}

			_, err = notedb.AddNoteItem(
				strings.Join(item.Tags, ","), item.Content)
			if err != nil {
				return 0, err
			}

			n++
			break
		}

		if !re.MatchString(str) {
			break
		}

		data := re.FindSubmatch([]byte(str))
		key := string(data[1])
		value := data[2]
		if !strings.HasPrefix(key, ITEM_PREFIX) {
			continue
		}

		key = trim_prefix(ITEM_PREFIX, key)
		if !strings.HasPrefix(key, othernotename) {
			continue
		}

		item := &Item{}
		if err := json.Unmarshal(value, item); err != nil {
			return 0, err
		}

		_, err = notedb.AddNoteItem(
			strings.Join(item.Tags, ","), item.Content)
		if err != nil {
			return 0, err
		}

		n++
	}

	err = notedb.db.Write(batch, nil)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func trim_prefix(p, s string) string {
	return regexp.MustCompile("^"+p).ReplaceAllString(s, "")
}
