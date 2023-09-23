package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tadvi/winc"
)

// Static variables
var FNT *winc.Font = winc.NewFont("SinsGold", 24, winc.DefaultFont.Style())
var TXTBOX_IMG string = "assets/spr_dialogue_base.png"

// Global variables
var dispMsg string = ""
var setMsg string = ""

func main() {
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
			tVar := 0

			for tVar < len(msg) {
				txt.SetText(txt.Text() + string(msg[tVar]))
				time.Sleep(50 * time.Millisecond)
				tVar += 1
			}
		}(dispMsg)

	})

	mainWindow.Center()
	mainWindow.Show()
	mainWindow.OnClose().Bind(wndOnClose)

	winc.RunMainLoop() // Must call to start event loop.
}

func wndOnClose(arg *winc.Event) {
	winc.Exit()
}

// formatString will take a string that's meant for GML and parse it into a format that
// GoLang is happy with.
func formatString(line string) (formattedLine string) {
	return strings.ReplaceAll(line, "#", "\n")
}

func parseString(line string) (speakerName string, retLine string, err error) {
	if strings.Index(line, " ") == -1 {
		return "", "", errors.New("invalid line")
	}

	command := line[:strings.Index(line, " ")]
	line = line[strings.Index(line, " ")+1:]
	switch command {
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
	default:
		return "", "", errors.New("invalid line")
	}

	return speakerName, line, nil
}

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
