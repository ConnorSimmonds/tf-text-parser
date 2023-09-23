package main

import (
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
