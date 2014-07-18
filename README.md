cnote
=====

A command line note app.

With cnote, you can conveniently manage note in command line. By the way, I installed a drop-down terminal [Yakuake](http://yakuake.kde.org), so I can call the terminal with a shortcut.

Cnote stores all data in a embedded database [goleveldb](https://github.com/syndtr/goleveldb), an implementation of the LevelDB key/value database in the Go programming language. The path of database files is ```~/.cnote/``` in *nix operating system and ```C:\Users\Administrator\.cnote\``` in Windows 7 for example.

Dependencies
------------

No. And it's platform independent, thanks for [golang](http://golang.org).


Usage
-----
    USAGE:
       cnote command [arguments...]
    
    COMMANDS:
       new          Create a new note
       del          Delete a note
       use          Select a note
       list, ls     List all notes
       
       add          add a note item
       rm           Remove a note item
       search, s    Search for
       tag, t       List items by tags
       
       dump         Dump whole database
       help, h      Shows a list of commands or help for one command


Examples
--------

    # Create a new note
    $ cnote new note1
    note "note1" created.
    current note: "note1".
    
    # List all notes
    $ cnote ls
    note1
    
    # add a note item
    $ cnote add tag1,tag2 "people mountain people sea"
    $ cnote add "tag1,tag2" "good good study, day day up"
    $ cnote add "tag3" "no zuo no nie"
    
    # List all tags
    $ cnote tag
    tag1
    tag2
    tag3
    
    # List items with a tag
    $ cnote tag tag2
    1: people mountain people sea
    2: good good study, day day up
    
    # Search items by keyword
    $ cnote search oo
    2: good good study, day day up
    
    # Dump whole database
    $ cnote dump
    "config":{"current_note_name":"note1"},
    "item_note1_000000001":{"itemid":"1","tags":["tag1","tag2"],"content":"people mountain people sea"},
    "item_note1_000000002":{"itemid":"2","tags":["tag1","tag2"],"content":"good good study, day day up"},
    "item_note1_000000003":{"itemid":"3","tags":["tag3"],"content":"no zuo no nie"},
    "note_note1":{"noteid":"note1","sum":3,"last_update":"2014-07-18 21:49:00 +0800 CST","last_id":3,"tags":{"tag1":{"1":true,"2":true},"tag2":{"1":true,"2":true},"tag3":{"3":true}}},
    

Copyright
--------

Copyright (c) 2014, Wei Shen (shenwei356@gmail.com)


[MIT License](https://github.com/shenwei356/cnote/blob/master/LICENSE)