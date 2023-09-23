package main

import (
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
var setMsg string = ""
var textSpeed int = 25
var dialogueFile *os.File

type Item struct {
	T       []string
	checked bool
}

func (item Item) Text() []string    { return item.T }
func (item *Item) SetText(s string) { item.T[0] = s }

func (item Item) Checked() bool            { return item.checked }
func (item *Item) SetChecked(checked bool) { item.checked = checked }
func (item Item) ImageIndex() int          { return 0 }

func main() {
	// Main window
	mainWindow := winc.NewForm(nil)
	mainWindow.SetSize(1200, 300)
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
	speakerTxt.SetText("Hello World")
	speakerTxt.SetSize(128, 32)
	speakerTxt.SetPos(16, 128)

	// Set up display field and button
	edt := winc.NewEdit(mainWindow)
	edt.SetPos(420, 24)
	edt.SetSize(384, 96)
	edt.SetText("")

	btn := winc.NewPushButton(mainWindow)
	btn.SetText("Display Line")
	btn.SetPos(420, 130)
	btn.SetSize(100, 40)
	btn.OnClick().Bind(func(e *winc.Event) {
		dispMsg = formatString(edt.Text())
		txt.SetText("")

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

			speakerTxt.SetText(speaker)

			// Now go and parse/display the rest of the string
			index := 0
			textSpeed = 50

			for index < len(msg) {
				if msg[index] == '[' {
					// start of a command, we need to parse through it and then apply the effects
					command := msg[index+1 : strings.Index(msg, "]")]
					parseCommand(command)

					msg = msg[strings.Index(msg, "]")+1:]
					index = 0
				}

				txt.SetText(txt.Text() + string(msg[index]))
				time.Sleep(time.Duration(textSpeed) * time.Millisecond)
				index += 1
			}
		}(dispMsg)

	})

	// Set up the line list for files
	lineList := winc.NewListView(mainWindow)
	lineList.AddColumn("Line", 160)
	lineList.AddColumn("Index", 60)
	lineList.SetCheckBoxes(false)
	lineList.SetPos(820, 0)

	loadBtn := winc.NewPushButton(mainWindow)
	loadBtn.SetPos(720, 200)
	loadBtn.SetSize(100, 40)
	loadBtn.SetText("Load Dialogue File")

	// Dialogue file load logic
	loadBtn.OnClick().Bind(func(e *winc.Event) {
		if filePath, ok := winc.ShowOpenFileDlg(mainWindow,
			"Select a dialogue file", "All files (*.*)|*.*", 0, ""); ok {

			itemList := parseDialogueFile(filePath)
			for _, item := range itemList {
				lineList.AddItem(item)
			}

		}
	})

	mainWindow.Center()
	mainWindow.Show()
	mainWindow.OnClose().Bind(wndOnClose)

	winc.RunMainLoop() // Must call to start event loop.
}

func wndOnClose(arg *winc.Event) {
	winc.Exit()
}

// parseDialogueFile parses in a dialogue file and returns an array of items to be
// added into the ListView
func parseDialogueFile(filePath string) (itemArray []Item) {
	itemArray = nil

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
