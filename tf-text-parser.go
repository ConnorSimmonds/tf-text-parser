package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tadvi/winc"
)

// Static variables
var FNT *winc.Font = winc.NewFont("SinsGold", 24, winc.DefaultFont.Style())
var TXTBOX_IMG string = "assets/spr_dialogue_base.png"
var TXTBOX *winc.Panel
var TXT *winc.Label

// Global variables
var dispMsg string = ""
var setMsg string = ""

func main() {
	mainWindow := winc.NewForm(nil)
	mainWindow.SetSize(1200, 300)
	mainWindow.SetText("Text Parser")

	TXTBOX := winc.NewPanel(mainWindow)
	TXTBOX.SetSize(384, 96)
	TXTBOX.SetPos(16, 24)

	txtBoxImg := winc.NewImageViewBox(TXTBOX)
	txtBoxImg.DrawImageFile(TXTBOX_IMG)
	TXT := winc.NewLabel(TXTBOX)
	TXT.SetText("hi")
	TXT.SetSize(374, 76)
	TXT.SetPos(5, 5)
	TXT.SetFont(FNT)

	edt := winc.NewEdit(mainWindow)
	edt.SetPos(420, 24)
	edt.SetSize(384, 96)
	edt.SetText("")

	btn := winc.NewPushButton(mainWindow)
	btn.SetText("Set Textbox Text")
	btn.SetPos(420, 130)
	btn.SetSize(100, 40)
	btn.OnClick().Bind(func(e *winc.Event) {
		dispMsg, err := strconv.Unquote("\"" + edt.Text() + "\"")
		if err != nil {
			fmt.Print(err)
			return
		}
		TXT.SetText("")

		go func(msg string) {
			tVar := 0
			for tVar < len(msg) {
				TXT.SetText(TXT.Text() + string(msg[tVar]))
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
