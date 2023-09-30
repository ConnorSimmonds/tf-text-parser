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
var textSpeed int = 25
var msgIndex int = -1
var dialogueFilePath string = ""
var textboxChannel = make(chan bool, 2)

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

func main() {
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

	saveBtn := winc.NewPushButton(mainWindow)
	saveBtn.SetPos(64, 280)
	saveBtn.SetSize(100, 40)
	saveBtn.SetText("Save Dialogue File")

	saveLineBtn := winc.NewPushButton(mainWindow)
	saveLineBtn.SetPos(264, 280)
	saveLineBtn.SetSize(100, 40)
	saveLineBtn.SetText("Save Dialogue Line")

	// Dialogue file load logic
	loadBtn.OnClick().Bind(func(e *winc.Event) {
		if filePath, ok := winc.ShowOpenFileDlg(mainWindow,
			"Select a dialogue file", "All files (*.*)|*.*", 0, ""); ok {
			lineList.DeleteAllItems()
			msgIndex = -1
			itemList := parseDialogueFile(filePath)
			for _, item := range itemList {
				lineList.AddItem(item)
			}
			dialogueFilePath = filePath
		}
	})

	// Save dialogue file (new)
	saveBtn.OnClick().Bind(func(e *winc.Event) {
		itemList := lineList.Items()
		var structList []Item

		for _, item := range itemList {
			var tItem = item.(*Item)
			structList = append(structList, *tItem)
		}

		saveDialogueFile(structList)
	})

	// Line replacement
	saveLineBtn.OnClick().Bind(func(e *winc.Event) {
		oldItm := lineList.SelectedItem()
		itm := &Item{[]string{edt.Text(), oldItm.Text()[1]}, true}
		lineList.InsertItem(itm, lineList.SelectedIndex())
		lineList.DeleteItem(oldItm)
		lineList.SetSelectedItem(itm)
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

	// Next Button Logic; this basically gets the next line and hits display
	nextBtn := winc.NewPushButton(mainWindow)
	nextBtn.SetPos(256, 130)
	nextBtn.SetSize(60, 40)
	nextBtn.SetText("Next")

	nextBtn.OnClick().Bind(func(e *winc.Event) {
		// get the next line: it's the next item on the selected item list
		// we also need to see if it's an "exit". if it is, stop playback
		selectedNext := lineList.SelectedIndex() + 1
		lineList.SetSelectedIndex(selectedNext)
		tItm := lineList.SelectedItem()
		fmt.Println(tItm)

		itmCont := tItm.Text()

		if itmCont[0] == "exit" {
			// we've reached the end!
			return
		}

		edt.SetText(itmCont[0])
		msgIndex, _ = strconv.Atoi(itmCont[1])

		// now, do the display code
		displayMessage(edt.Text(), txt, speakerTxt)
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
	textBox.SetText("")

	go func(msg string) {
		// Parse and display the string
		// The general structure of most lines is dia x line
		// Dia is the speaker marker.
		speaker, msg, err := parseString(msg)

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

		for index < len(msg) {
			// check to see if we've told the goroutine to stop
			select {
			case <-textboxChannel:
				fmt.Println("Stop early")
				return
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
		textboxChannel <- true
	}(dispMsg)
}

// saveDialogueFile takes a list of items from the lineList and saves them to the loaded in file
// Right now this doesn't support adding in extra lines.
func saveDialogueFile(diaList []Item) {
	if dialogueFilePath == "" {
		return
	}

	// parse the file in
	fileString, err := os.ReadFile(dialogueFilePath)
	if err != nil {
		panic(err)
	}

	baseFileLines := strings.Split(string(fileString), "\n")

	// TODO: how do we handle new lines added in? We don't have this functionality right now, but I'd like to add it in as a way to have an AIO tool
	for _, diaItem := range diaList {
		ind, _ := strconv.Atoi(diaItem.Text()[1])
		line := diaItem.Text()[0]

		if !diaItem.checked {
			continue
		}

		baseFileLines[ind] = line
	}

	output := strings.Join(baseFileLines, "\n")
	err = os.WriteFile(dialogueFilePath, []byte(output), 0644)
	if err != nil {
		panic(err)
	}
}

// parseDialogueFile parses in a dialogue file and returns an array of items to be
// added into the ListView
func parseDialogueFile(filePath string) []*Item {
	var itemArray []*Item

	// parse the file in
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	dialogueScanner := bufio.NewScanner(file)
	index := -1

	// go through the file and add each line into the item
	for dialogueScanner.Scan() {
		index += 1
		txt := dialogueScanner.Text()

		if strings.Index(txt, "dia") != 0 && strings.Index(txt, "exit") != 0 {
			continue
		}
		itm := &Item{[]string{txt, strconv.Itoa(index)}, false}
		itemArray = append(itemArray, itm)

	}

	return itemArray
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
