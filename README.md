cnote
=====

A **platform independent** command line note app.

Cnote supports **backup** and **restoring** from backup, you can also **import** notes from backup of others.

With cnote, you can conveniently manage note in command line. You can install a drop-down terminal like [Yakuake](http://yakuake.kde.org), so you can instantly call the terminal with a shortcut.

Since cnote is a command line tool, it is suitabl suitable for plain text of tens of words. 

Cnote stores all data in a embedded database [goleveldb](https://github.com/syndtr/goleveldb), an implementation of the LevelDB key/value database in the Go programming language. The path of database files is ```~/.cnote/``` in *nix operating system and ```C:\Users\Administrator\.cnote\``` in Windows 7 for example.

Dependencies
------------

No. Thanks for [golang](http://golang.org) and [goleveldb](https://github.com/syndtr/goleveldb).

Download
--------

[**Recommanded**] To compile with the newest source code, please use [gobuild - Cross-Platform Go Project Compiler](http://gobuild.io/download/github.com/shenwei356/cnote). It's simple and fast!

Download lastest release from [release page](https://github.com/shenwei356/cnote/releases) or from my personal site: [cnoteV1.1.zip](http://blog.shenwei.me/?wpdmact=process&did=Ni5ob3RsaW5r).


Usage
-----

    USAGE:
       cnote command [arguments...]
    
    COMMANDS:
       new          Create a new note
       del          Delete a note
       use          Select a note
       list, ls     List all notes
       
       add          Add a note item
       rm           Remove a note item
       tag, t       List items by tags. List all tags if no arguments given
       search, s    Search items with regular expression
       
       dump         Dump whole database, for backup or transfer
       wipe         Attention! Wipe whole database
       restore      Wipe whole database, and restore from dumpped file
       import       Import note items from dumpped data
       
       help, h      Shows a list of commands or help for one command


Examples
--------

    ############### Create a new note ###############

    $ cnote new fruit
    note "fruit" created.
    current note: "fruit".
    $ cnote new people
    note "people" created.
    current note: "people". 

    ############### List all notes ###############

    $ cnote ls
    note: fruit     (#. of items: 0, last update: 2014-07-20 04:07:00 +0800 CST).
    note: people    (#. of items: 0, last update: 2014-07-20 04:07:00 +0800 CST). (current note)

    ############### Choose note fruit ###############

    $ cnote use fruit
    current note: "fruit" (last update: 2014-07-20 04:07:00 +0800 CST).

    ############### Delete a new note ###############
    
    $ cnote del test

    ###########################################################################    
    
    ############### add note item ###############

    $ cnote add red,green apple
    item: 1 (tags: [red green])     apple
    $ cnote add green,yellow pear
    item: 2 (tags: [green yellow])  pear
    $ cnote add yellow banana
    item: 3 (tags: [yellow])        banana

    ############### Show all tags ###############

    $ cnote tag
    tag: green      (#. of items: 2).
    tag: red        (#. of items: 1).
    tag: yellow     (#. of items: 2).

    ############### Show items by tag ###############

    $ cnote tag yellow
    item: 2 (tags: [green yellow])  pear
    item: 3 (tags: [yellow])        banana

    ############### Search items by regrexp  ###############

    $ cnote s ea
    item: 2 (tags: [green yellow])  pear

    ############### Show all items, just search with .  ###############

    $ cnote s .
    item: 1 (tags: [red green])     apple
    item: 2 (tags: [green yellow])  pear
    item: 3 (tags: [yellow])        banana
    
    ############### remove a note item ###############

    $ cnote s .
    item: 1 (tags: [red green])     apple
    item: 2 (tags: [green yellow])  pear
    item: 3 (tags: [yellow])        banana
    $ cnote rm 2 
    $ cnote s .
    item: 1 (tags: [red green])     apple
    item: 3 (tags: [yellow])        banana

    ###########################################################################

    ############### Dump database for backup  ###############

    $ cnote dump
    config  {"current_note_name":"fruit"}
    item_fruit_000000001    {"itemid":"1","tags":["red","green"],"content":"apple"}
    item_fruit_000000002    {"itemid":"2","tags":["green","yellow"],"content":"pear"}
    item_fruit_000000003    {"itemid":"3","tags":["yellow"],"content":"banana"}
    note_fruit      {"noteid":"fruit","sum":3,"last_update":"2014-07-20 04:13:00 +0800 CST","last_id":3,"tags":{"green":{"1":true,"2":true},"red":{"1":true},"yellow":{"2":true,"3":true}}}
    note_people     {"noteid":"people","sum":0,"last_update":"2014-07-20 04:07:00 +0800 CST","last_id":0,"tags":{}}

    $ cnote dump > dumpdata 

    ############### Wipe whole database, and restore from dumpped file  ###############

    $ cnote restore dumpdata 
    Attention, it will clear all the data. type "yes" to continue:yes

    ############### Import note items from dumpped data  ###############

    $ cnote import fruit fruit dumpdata 
    3 items imported into note "fruit".
    $ cnote dump
    config  {"current_note_name":"fruit"}
    item_fruit_000000001    {"itemid":"1","tags":["red","green"],"content":"apple"}
    item_fruit_000000002    {"itemid":"2","tags":["green","yellow"],"content":"pear"}
    item_fruit_000000003    {"itemid":"3","tags":["yellow"],"content":"banana"}
    item_fruit_000000004    {"itemid":"4","tags":["red","green"],"content":"apple"}
    item_fruit_000000005    {"itemid":"5","tags":["green","yellow"],"content":"pear"}
    item_fruit_000000006    {"itemid":"6","tags":["yellow"],"content":"banana"}
    note_fruit      {"noteid":"fruit","sum":6,"last_update":"2014-07-20 04:22:00 +0800 CST","last_id":6,"tags":{"green":{"1":true,"2":true,"4":true,"5":true},"red":{"1":true,"4":true},"yellow":{"2":true,"3":true,"5":true,"6":true}}}
    note_people     {"noteid":"people","sum":0,"last_update":"2014-07-20 04:07:00 +0800 CST","last_id":0,"tags":{}}


Copyright
--------

Copyright (c) 2014, Wei Shen (shenwei356@gmail.com)


[MIT License](https://github.com/shenwei356/cnote/blob/master/LICENSE)