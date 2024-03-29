package ui

import (
	"log"
	"time"

	"github.com/dmars8047/broterm/internal/state"
	"github.com/dmars8047/idamlib/idam"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const HOME_PAGE PageSlug = "home"

type HomePage struct {
	userAuthClient   *idam.UserAuthClient
	currentThemeCode string
}

func NewHomePage(userAuthClient *idam.UserAuthClient) *HomePage {
	return &HomePage{
		userAuthClient:   userAuthClient,
		currentThemeCode: "NOT_SET",
	}
}

func (page *HomePage) Setup(app *tview.Application, appContext *state.ApplicationContext, nav *PageNavigator) {

	grid := tview.NewGrid()

	grid.SetRows(4, 8, 8, 0).
		SetColumns(0, 31, 39, 0)

	logoBro := tview.NewTextView()
	logoBro.SetTextAlign(tview.AlignLeft)
	logoBro.SetText(
		`BBBBBBB\                      
BB  __BB\                     
BB |  BB | RRRRRR\   OOOOOO\  
BBBBBBB\ |RR  __RR\ OO  __OO\ 
BB  __BB\ RR |  \__|OO /  OO |
BB |  BB |RR |      OO |  OO |
BBBBBBB  |RR |      \OOOOOO  |
\_______/ \__|       \______/ `)

	logoChat := tview.NewTextView()
	logoChat.SetTextAlign(tview.AlignLeft)
	logoChat.SetText(
		` CCCCCC\  HH\                  TT\
CC  __CC\ HH |                 TT |
CC /  \__|HHHHHHH\   AAAAAA\ TTTTTT\
CC |      HH  __HH\  \____AA\\_TT  _|
CC |      HH |  HH | AAAAAAA | TT |
CC |  CC\ HH |  HH |AA  __AA | TT |TT\
\CCCCCC  |HH |  HH |\AAAAAAA | \TTTT  |
 \______/ \__|  \__| \_______|  \____/`)

	brosButton := tview.NewButton("Bros")

	brosButton.SetSelectedFunc(func() {
		nav.NavigateTo(FRIENDS_LIST_PAGE, nil)
	})

	chatButton := tview.NewButton("Chat")

	chatButton.SetSelectedFunc(func() {
		nav.NavigateTo(ROOM_LIST_PAGE, nil)
	})

	logoutButton := tview.NewButton("Logout")

	logoutButton.SetSelectedFunc(func() {
		accessToken, ok := appContext.GetAccessToken()

		if !ok {
			log.Printf("Valid user authentication information not found. Redirecting to login page.")
			nav.NavigateTo(LOGIN_PAGE, nil)
			return
		}

		err := page.userAuthClient.Logout(accessToken)

		if err != nil {
			nav.AlertFatal(app, "home:menu:alert:err", err.Error())
			return
		}

		appContext.CancelUserSession()

		nav.NavigateTo(WELCOME_PAGE, nil)
	})

	buttonGrid := tview.NewGrid()

	tvInstructions := tview.NewTextView().SetTextAlign(tview.AlignCenter)

	logoutButton.SetFocusFunc(func() {
		tvInstructions.SetText("Sign out of your account.")
	})

	chatButton.SetFocusFunc(func() {
		tvInstructions.SetText("Chat in a room or find one to join.")
	})

	brosButton.SetFocusFunc(func() {
		tvInstructions.SetText("Talk to your Bros or find new ones!")
	})

	buttonGrid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

		goRight := func() {
			if brosButton.HasFocus() {
				app.SetFocus(chatButton)
			} else if chatButton.HasFocus() {
				app.SetFocus(logoutButton)
			} else if logoutButton.HasFocus() {
				app.SetFocus(brosButton)
			}
		}

		goLeft := func() {
			if logoutButton.HasFocus() {
				app.SetFocus(chatButton)
			} else if chatButton.HasFocus() {
				app.SetFocus(brosButton)
			} else if brosButton.HasFocus() {
				app.SetFocus(logoutButton)
			}
		}

		// vim movement keys
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 'l':
				goRight()
			case 'h':
				goLeft()
			}
		} else if event.Key() == tcell.KeyTab || event.Key() == tcell.KeyRight {
			goRight()
		} else if event.Key() == tcell.KeyBacktab || event.Key() == tcell.KeyLeft {
			goLeft()
		}

		return event
	})

	buttonGrid.SetRows(3, 1, 1).
		SetColumns(0, 1, 0, 1, 0)

	buttonGrid.AddItem(brosButton, 0, 0, 1, 1, 0, 0, true).
		AddItem(chatButton, 0, 2, 1, 1, 0, 0, true).
		AddItem(logoutButton, 0, 4, 1, 1, 0, 0, true).
		AddItem(tvInstructions, 2, 0, 1, 5, 0, 0, false)

	grid.AddItem(logoBro, 1, 1, 1, 1, 0, 0, false).
		AddItem(logoChat, 1, 2, 1, 1, 0, 0, false).
		AddItem(buttonGrid, 2, 1, 1, 2, 0, 0, true)

	// Apply all colors and styles
	applyTheme := func() {
		theme := appContext.GetTheme()

		if theme.Code != page.currentThemeCode {
			page.currentThemeCode = theme.Code
			grid.SetBackgroundColor(theme.BackgroundColor)

			logoBro.SetBackgroundColor(theme.BackgroundColor)
			logoBro.SetTextColor(tcell.ColorWhite)
			logoChat.SetBackgroundColor(theme.BackgroundColor)
			logoChat.SetBackgroundColor(theme.BackgroundColor)
			logoChat.SetTextColor(theme.HighlightColor)

			brosButton.SetActivatedStyle(theme.ActivatedButtonStyle)
			brosButton.SetStyle(theme.ButtonStyle)

			chatButton.SetActivatedStyle(theme.ActivatedButtonStyle)
			chatButton.SetStyle(theme.ButtonStyle)

			logoutButton.SetActivatedStyle(theme.ActivatedButtonStyle)
			logoutButton.SetStyle(theme.ButtonStyle)

			tvInstructions.SetBackgroundColor(theme.BackgroundColor)
			tvInstructions.SetTextColor(theme.ForgroundColor)
		}
	}

	applyTheme()

	nav.Register(HOME_PAGE, grid, true, false,
		func(_ interface{}) {
			applyTheme()
			page.onPageLoad(appContext, nav)
		}, func() {
			page.onPageClose()
		})
}

func (page *HomePage) onPageLoad(appContext *state.ApplicationContext, nav *PageNavigator) {
	// Make sure the session is still valid
	if appContext.GetUserAuth().TokenExpiration.Before(time.Now()) {
		appContext.CancelUserSession()
		nav.NavigateTo(LOGIN_PAGE, nil)
	}
}

func (page *HomePage) onPageClose() {
	// Nothing to do here
}
