package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/tadvi/winc"
)

// Static variables
var FNT *winc.Font = winc.NewFont("SinsGold", 24, winc.DefaultFont.Style())
var TXTBOX_IMG string = "assets/spr_dialogue_base.png"
var BASE_SPEED int = 50

// Global variables
var dispMsg string = ""
var finalMsg string = ""
var textSpeed int = 25
var msgIndex int = -1
var dialogueFilePath string = ""
var textboxChannel = make(chan bool, 1)
var finishedChannel = make(chan bool, 1)

// Item Functions
type Item struct {
	T       []string
	checked bool
}

func (item Item) Text() []string    { return item.T }
func (item *Item) SetText(s string) { item.T[0] = s }

func (item Item) Checked() bool            { return item.checked }
func (item *Item) SetChecked(checked bool) { item.checked = checked }
func (item Item) ImageIndex() int          { return 0 }
func (item Item) Index() string            { return item.T[1] }
func (item *Item) SetIndex(index string)   { item.T[1] = index }

// ConvItem Functions
type ConvItem struct {
	T       []string
	checked bool
	conv    []*Item
}

func (item ConvItem) Text() []string    { return item.T }
func (item *ConvItem) SetText(s string) { item.T[0] = s }

func (item ConvItem) Checked() bool                 { return item.checked }
func (item *ConvItem) SetChecked(checked bool)      { item.checked = checked }
func (item ConvItem) ImageIndex() int               { return 0 }
func (item ConvItem) Index() string                 { return item.T[1] }
func (item *ConvItem) SetIndex(index string)        { item.T[1] = index }
func (item ConvItem) Conversation() []*Item         { return item.conv }
func (item *ConvItem) SetConversation(conv []*Item) { item.conv = conv }

func main() {
	// Preseed the finishedChannel so that we can display a line
	finishedChannel <- true

	// Main window
	mainWindow := winc.NewForm(nil)
	mainWindow.SetSize(960, 460)
	mainWindow.SetText("Text Parser")

	// Set up textbox
	txtbox := winc.NewPanel(mainWindow)
	txtbox.SetSize(384, 96)
	txtbox.SetPos(16, 24)

	txtBoxImg := winc.NewImageViewBox(txtbox)
	txtBoxImg.DrawImageFile(TXTBOX_IMG)
	txt := winc.NewLabel(txtbox)
	txt.SetText("")
	txt.SetSize(374, 76)
	txt.SetPos(5, 5)
	txt.SetFont(FNT)

	// Set up the speaker label
	speakerTxt := winc.NewLabel(mainWindow)
	speakerTxt.SetText("")
	speakerTxt.SetSize(128, 32)
	speakerTxt.SetPos(16, 128)

	// Set up display field and button
	edt := winc.NewEdit(mainWindow)
	edt.SetPos(16, 172)
	edt.SetSize(384, 96)
	edt.SetText("")

	btn := winc.NewPushButton(mainWindow)
	btn.SetText("Display Line")
	btn.SetPos(128, 130)
	btn.SetSize(100, 40)
	btn.OnClick().Bind(func(e *winc.Event) {
		if strings.Index(edt.Text(), "dia") != 0 {
			return
		}
		displayMessage(edt.Text(), txt, speakerTxt)
	})

	// Set up the line list for files
	conversationList := winc.NewListView(mainWindow)
	conversationList.AddColumn("Conversation", 160)
	conversationList.SetCheckBoxes(false)
	conversationList.SetPos(420, 0)

	lineList := winc.NewListView(mainWindow)
	lineList.AddColumn("Line", 160)
	lineList.AddColumn("Index", 60)
	lineList.SetCheckBoxes(false)
	lineList.EnableSingleSelect(true)
	lineList.SetPos(640, 0)

	loadBtn := winc.NewPushButton(mainWindow)
	loadBtn.SetPos(164, 280)
	loadBtn.SetSize(100, 40)
	loadBtn.SetText("Load Dialogue File")
	// Dialogue file load logic
	loadBtn.OnClick().Bind(func(e *winc.Event) {
		if filePath, ok := winc.ShowOpenFileDlg(mainWindow,
			"Select a dialogue file", "All files (*.*)|*.*", 0, ""); ok {
			lineList.DeleteAllItems()
			msgIndex = -1
			itemList := parseDialogueFile(filePath)
			for _, item := range itemList {
				conversationList.AddItem(item)
			}
			dialogueFilePath = filePath
		}
	})

	saveBtn := winc.NewPushButton(mainWindow)
	saveBtn.SetPos(64, 280)
	saveBtn.SetSize(100, 40)
	saveBtn.SetText("Save Dialogue File")

	saveConvBtn := winc.NewPushButton(mainWindow)
	saveConvBtn.SetPos(264, 280)
	saveConvBtn.SetSize(100, 40)
	saveConvBtn.SetText("Save Conversation")
	saveConvBtn.OnClick().Bind(func(e *winc.Event) {
		convItem := conversationList.SelectedItem().(*ConvItem)
		lineArray := make([]*Item, lineList.ItemCount())

		for _, line := range lineList.Items() {
			lineItem := line.(*Item)
			lineIndex, _ := strconv.Atoi(lineItem.Index())
			lineArray[lineIndex] = lineItem
		}

		convItem.SetConversation(lineArray)
		conversationList.UpdateItem(convItem)
	})

	// Save dialogue file (new)
	saveBtn.OnClick().Bind(func(e *winc.Event) {
		itemList := conversationList.Items()
		structList := make([]ConvItem, conversationList.ItemCount())

		for _, item := range itemList {
			var tItem = item.(*ConvItem)
			index, _ := strconv.Atoi(tItem.Text()[0])
			structList[index] = *tItem
		}

		saveDialogueFile(structList)
	})

	// Dialogue list line click logic
	lineList.OnClick().Bind(func(e *winc.Event) {
		itm := lineList.SelectedItem()
		if itm == nil {
			return
		}

		itmCont := itm.Text()
		edt.SetText(itmCont[0])
		msgIndex, _ = strconv.Atoi(itmCont[1])
	})

	// Conversation list line click logic
	conversationList.OnClick().Bind(func(e *winc.Event) {
		itm := conversationList.SelectedItem()
		if itm == nil {
			return
		}

		var convItem = itm.(*ConvItem)
		lineList.DeleteAllItems()
		for _, itm := range convItem.Conversation() {
			lineList.AddItem(itm)
		}

		lineList.SetSelectedIndex(0)
		lineItm := lineList.SelectedItem().(*Item)
		itmCont := lineItm.Text()
		edt.SetText(itmCont[0])

		msgIndex, _ = strconv.Atoi(itmCont[1])
	})

	// Next Button Logic; this basically gets the next line and hits display
	nextBtn := winc.NewPushButton(mainWindow)
	nextBtn.SetPos(256, 130)
	nextBtn.SetSize(60, 40)
	nextBtn.SetText("Next")

	nextBtn.OnClick().Bind(func(e *winc.Event) {
		// get the next line: it's the next item on the selected item list
		// we also need to see if it's an "exit". if it is, stop playback
		// if we press next on an "exit" we dont want to do anything
		if lineList.SelectedItem().Text()[0] == "exit" {
			lineList.SetSelectedIndex(0)
			return
		}

		for true {
			selectedNext := lineList.SelectedIndex() + 1
			lineList.SetSelectedIndex(selectedNext)
			tItm := lineList.SelectedItem()
			itmCont := tItm.Text()

			if itmCont[0] == "exit" {
				// we've reached the end!
				lineList.SetSelectedIndex(0)
				tItm := lineList.SelectedItem()
				itmCont = tItm.Text()
			}

			// check to see if this is a dialogue line, if not, get the next line....
			if strings.Index(itmCont[0], "dia") == 0 {
				edt.SetText(itmCont[0])
				msgIndex, _ = strconv.Atoi(itmCont[1])
				break
			}
		}

		// now, do the display code
		displayMessage(edt.Text(), txt, speakerTxt)
	})

	// delete/add line buttons
	addLineBtn := winc.NewPushButton(mainWindow)
	addLineBtn.SetPos(64, 320)
	addLineBtn.SetSize(100, 40)
	addLineBtn.SetText("Add Line")
	addLineBtn.OnClick().Bind(func(e *winc.Event) {
		// adds a new line
		itm := &Item{[]string{"dia -1 ", "0"}, false}
		lineList.InsertItem(itm, lineList.SelectedIndex()+1)
		fixLineListIndexes(lineList)
	})

	removeLineBtn := winc.NewPushButton(mainWindow)
	removeLineBtn.SetPos(164, 320)
	removeLineBtn.SetSize(100, 40)
	removeLineBtn.SetText("Remove Line")
	removeLineBtn.OnClick().Bind(func(e *winc.Event) {
		//removes a line
		lineList.DeleteItem(lineList.SelectedItem())
		// correct the indexes
		fixLineListIndexes(lineList)
	})

	upLineBtn := winc.NewPushButton(mainWindow)
	upLineBtn.SetPos(24, 300)
	upLineBtn.SetSize(40, 40)
	upLineBtn.SetText("^")
	upLineBtn.OnClick().Bind(func(e *winc.Event) {
		if lineList.SelectedIndex() <= 0 {
			return
		}

		// move the current line down
		tItem := lineList.SelectedItem().(*Item)
		newIndex, _ := strconv.Atoi(tItem.Index())
		tItem.SetIndex(strconv.Itoa(newIndex - 2))
		insertIndex := lineList.SelectedIndex() - 1
		lineList.DeleteItem(lineList.SelectedItem())
		lineList.InsertItem(tItem, insertIndex)
		lineList.SetSelectedIndex(insertIndex)
		// correct the indexes
		fixLineListIndexes(lineList)
	})

	downLineBtn := winc.NewPushButton(mainWindow)
	downLineBtn.SetPos(24, 340)
	downLineBtn.SetSize(40, 40)
	downLineBtn.SetText("v")
	downLineBtn.OnClick().Bind(func(e *winc.Event) {
		// move the current line down
		if lineList.SelectedIndex() >= lineList.ItemCount() {
			return
		}

		tItem := lineList.SelectedItem().(*Item)
		newIndex, _ := strconv.Atoi(tItem.Index())
		tItem.SetIndex(strconv.Itoa(newIndex + 2))
		insertIndex := lineList.SelectedIndex() + 1
		lineList.DeleteItem(lineList.SelectedItem())
		lineList.InsertItem(tItem, insertIndex)
		lineList.SetSelectedIndex(insertIndex)
		// correct the indexes
		fixLineListIndexes(lineList)
	})

	// Line replacement
	saveLineBtn := winc.NewPushButton(mainWindow)
	saveLineBtn.SetPos(264, 320)
	saveLineBtn.SetSize(100, 40)
	saveLineBtn.SetText("Save Dialogue Line")
	saveLineBtn.OnClick().Bind(func(e *winc.Event) {
		itm := lineList.SelectedItem().(*Item)
		itm.SetChecked(true)
		itm.SetText(edt.Text())
		lineList.UpdateItem(itm)
	})

	mainWindow.Center()
	mainWindow.Show()
	mainWindow.OnClose().Bind(wndOnClose)

	winc.RunMainLoop() // Must call to start event loop.
}

func wndOnClose(arg *winc.Event) {
	winc.Exit()
}

// displayMessage sets up everything to be displayed. This is reusable so we can call it from Next.
func displayMessage(rawLine string, textBox *winc.Label, speakerBox *winc.Label) {
	dispMsg = formatString(rawLine)

	if textBox.Text() != finalMsg {
		textboxChannel <- true
	}
	<-finishedChannel

	textBox.SetText("")

	go func(inpMsg string) {
		// Parse and display the string
		// The general structure of most lines is dia x line
		// Dia is the speaker marker.
		speaker, msg, err := parseString(inpMsg)
		finalMsg = msg

		if err != nil {
			fmt.Print(err)
			return
		}

		// it's a recognized command but not a dialogue command
		if speaker == "" && msg == "" {
			return
		}

		speakerBox.SetText(speaker)

		// Now go and parse/display the rest of the string
		index := 0
		textSpeed = 50

	displayLoop:
		for index < len(msg) {
			// check to see if we've told the goroutine to stop
			select {
			case <-textboxChannel:
				break displayLoop
			default:
				if msg[index] == '[' {
					// start of a command, we need to parse through it and then apply the effects
					command := msg[index+1 : strings.Index(msg, "]")]
					parseCommand(command)

					msg = msg[strings.Index(msg, "]")+1:]
					index = 0
				}
				textBox.SetText(textBox.Text() + string(msg[index]))
				time.Sleep(time.Duration(textSpeed) * time.Millisecond)
				index += 1
			}
		}
		finishedChannel <- true
	}(dispMsg)
}

// saveDialogueFile takes a list of items from the lineList and saves them to the loaded in file
// Right now this doesn't support adding in extra lines.
func saveDialogueFile(convList []ConvItem) {
	if dialogueFilePath == "" {
		return
	}

	var newFile []string

	for _, convItem := range convList {
		convos := convItem.Conversation()
		for _, diaItem := range convos {
			//ind, _ := strconv.Atoi(diaItem.Text()[1])
			line := diaItem.Text()[0]
			newFile = append(newFile, line)
		}

	}

	output := strings.Join(newFile, "\n")
	fmt.Println(output)
	err := os.WriteFile(dialogueFilePath, []byte(output), 0644)
	if err != nil {
		panic(err)
	}
}

// parseDialogueFile parses in a dialogue file and returns an array of items to be
// added into the ListView
func parseDialogueFile(filePath string) []*ConvItem {
	var itemArray []*Item
	var convArray []*ConvItem

	// parse the file in
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	dialogueScanner := bufio.NewScanner(file)
	index := -1
	convIndex := 0

	// go through the file and add each line into the item
	for dialogueScanner.Scan() {
		index += 1
		txt := dialogueScanner.Text()

		itm := &Item{[]string{txt, strconv.Itoa(index)}, false}
		itemArray = append(itemArray, itm)

		if txt == "exit" {
			convItem := &ConvItem{[]string{strconv.Itoa(convIndex)}, false, itemArray}
			convIndex += 1
			convArray = append(convArray, convItem)
			itemArray = []*Item{}
			index = -1
		}
	}

	return convArray
}

/*
	Helper Functions for Parsing Strings
*/

// formatString will take a string that's meant for GML and parse it into a format that
// GoLang is happy with.
func formatString(line string) (formattedLine string) {
	return strings.ReplaceAll(line, "#", "\n")
}

// parseString parses the string, as our line structure is [command] [value..value_n].
// Ergo, a dialogue line is dia [speakerId] [line]
func parseString(line string) (speakerName string, retLine string, err error) {
	if strings.Index(line, " ") == -1 {
		return "", "", errors.New("invalid line")
	}

	command := line[:strings.Index(line, " ")]
	line = line[strings.Index(line, " ")+1:]
	switch command {
	case "dia":
		// get the speaker id
		speakerId, err := strconv.ParseInt(line[:strings.Index(line, " ")], 10, 32)

		if err != nil {
			fmt.Println(err)
			return "", "", errors.New("invalid line")
		}

		speakerName = getSpeaker(speakerId)
		line = line[strings.Index(line, " ")+1:]
		break
	case "set_pos":
		fallthrough
	case "create":
		fallthrough
	case "em":
		fallthrough
	case "music":
		fallthrough
	case "move":
		fallthrough
	case "del":
		fallthrough
	case "name":
		fallthrough
	case "speed":
		fallthrough
	case "pause":
		fallthrough
	case "path":
		fallthrough
	case "area_show":
		fallthrough
	case "set_zoom":
		fallthrough
	case "hide_dia":
		fallthrough
	case "focus":
		fallthrough
	case "tut_setup":
		fallthrough
	case "exp_setup":
		fallthrough
	case "abi":
		fallthrough
	case "line":
		fallthrough
	case "file":
		fallthrough
	case "exit":
		fallthrough
	case "battle":
		fallthrough
	case "screensh":
		fallthrough
	case "slow":
		fallthrough
	case "shake":
		fallthrough
	case "to_menu":
		return "", "", nil
	default:
		return "", "", errors.New("invalid line")
	}

	return speakerName, line, nil
}

// getSpeaker returns the speaker string. Sourced from the game
func getSpeaker(speakerId int64) string {
	switch speakerId {
	case -1:
		fallthrough
	case 0:
		return ""
	case 1:
		return "Eduardo"
	case 2:
		return "Violet"
	case 3:
		return "'Kat'"
	default:
		return "N/A"
	}
}

// parseCommand takes in a command and parses it.
// This is massively truncated from the main one.
func parseCommand(commandString string) {
	command := commandString[:strings.Index(commandString, ":")]
	value := ""

	if strings.Index(commandString, ":")+1 != -1 {
		value = commandString[(strings.Index(commandString, ":") + 1):]
	}

	switch command {
	case "P":
		pause, _ := strconv.Atoi(value)
		time.Sleep(time.Duration((int(float64(BASE_SPEED)*0.75))*pause) * time.Millisecond)
	case "S":
		speedMod, _ := strconv.ParseFloat(value, 10)
		textSpeed = int((float64(BASE_SPEED)) / speedMod)
		break
	default:
		fmt.Println(command + ":" + value)
		break
	}
}

func fixLineListIndexes(list *winc.ListView) {
	// goes through the list and fixes up the indexes
	index := 0
	oldSelect := list.SelectedIndex()
	for index < list.ItemCount() {
		list.SetSelectedIndex(index)
		item := list.SelectedItem().(*Item)
		item.SetIndex(strconv.Itoa(index))
		list.UpdateItem(item)
		index += 1
	}
	list.SetSelectedIndex(oldSelect)
}
