package ui

import (
	"fmt"
	"time"

	"github.com/dmars8047/brochat-service/pkg/chat"
	"github.com/dmars8047/broterm/internal/state"
	"github.com/dmars8047/idam-service/pkg/idam"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type HomeModule struct {
	appState       state.ApplicationState
	userAuthClient *idam.UserAuthClient
	brochatClient  *chat.BroChatUserClient
	pageNav        *PageNavigator
	app            *tview.Application
}

func NewHomeModule(userAuthClient *idam.UserAuthClient,
	application *tview.Application,
	pageNavigator *PageNavigator,
	brochatClient *chat.BroChatUserClient,
	appState state.ApplicationState) *HomeModule {
	return &HomeModule{
		appState:       appState,
		brochatClient:  brochatClient,
		userAuthClient: userAuthClient,
		app:            application,
		pageNav:        pageNavigator,
	}
}

func (mod *HomeModule) SetupHomePages() {
	mod.setupMenuPage()
	mod.setupFriendListPage()
	mod.setupFindAFriendPage()
}

func (mod *HomeModule) setupMenuPage() {
	grid := tview.NewGrid()
	grid.SetBackgroundColor(DefaultBackgroundColor)

	grid.SetRows(4, 8, 8, 0).
		SetColumns(0, 31, 39, 0)

	logoBro := tview.NewTextView()
	logoBro.SetTextAlign(tview.AlignLeft).
		SetBackgroundColor(DefaultBackgroundColor)
	logoBro.SetTextColor(tcell.ColorWhite)
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
	logoChat.SetBackgroundColor(DefaultBackgroundColor)
	logoChat.SetTextColor(BroChatYellowColor)
	logoChat.SetText(
		` CCCCCC\  HH\                  TT\
CC  __CC\ HH |                 TT |
CC /  \__|HHHHHHH\   AAAAAA\ TTTTTT\
CC |      HH  __HH\  \____AA\\_TT  _|
CC |      HH |  HH | AAAAAAA | TT |
CC |  CC\ HH |  HH |AA  __AA | TT |TT\
\CCCCCC  |HH |  HH |\AAAAAAA | \TTTT  |
 \______/ \__|  \__| \_______|  \____/`)

	brosButton := tview.NewButton("Bros").
		SetActivatedStyle(ActivatedButtonStyle).
		SetStyle(ButtonStyle)

	brosButton.SetSelectedFunc(func() {
		mod.pageNav.NavigateTo(HOME_FRIENDS_LIST_PAGE)
	})

	chatButton := tview.NewButton("Chat").
		SetActivatedStyle(ActivatedButtonStyle).
		SetStyle(ButtonStyle)

	chatButton.SetSelectedFunc(func() {
		Alert(mod.pageNav.Pages, "home:menu:alert:info", "You pressed the chat button")
	})

	settingsButton := tview.NewButton("Settings").
		SetActivatedStyle(ActivatedButtonStyle).
		SetStyle(ButtonStyle)

	settingsButton.SetSelectedFunc(func() {
		Alert(mod.pageNav.Pages, "home:menu:alert:info", "You pressed the settings button")
	})

	logoutButton := tview.NewButton("Logout").
		SetActivatedStyle(ActivatedButtonStyle).
		SetStyle(ButtonStyle)

	logoutButton.SetSelectedFunc(func() {
		session, ok := state.Get[state.UserSession](mod.appState, state.UserSessionProp)

		if !ok {
			AlertFatal(mod.app, mod.pageNav.Pages, "home:menu:alert:err", "Application State Error - Could not get user session.")
		}

		err := mod.userAuthClient.Logout(session.Auth.AccessToken)

		if err != nil {
			AlertFatal(mod.app, mod.pageNav.Pages, "home:menu:alert:err", err.Error())
			return
		}

		state.Set(mod.appState, state.UserSessionProp, nil)

		mod.pageNav.NavigateTo(WELCOME_PAGE)
	})

	buttonGrid := tview.NewGrid()

	tvInstructions := tview.NewTextView().SetTextAlign(tview.AlignCenter)
	tvInstructions.SetBackgroundColor(DefaultBackgroundColor)
	tvInstructions.SetTextColor(tcell.ColorWhite)

	logoutButton.SetFocusFunc(func() {
		tvInstructions.SetText("Sign out of your account.")
	})

	settingsButton.SetFocusFunc(func() {
		tvInstructions.SetText("Change your account settings.")
	})

	chatButton.SetFocusFunc(func() {
		tvInstructions.SetText("Chat in a server or find one to join.")
	})

	brosButton.SetFocusFunc(func() {
		tvInstructions.SetText("Talk to your Bros or find new ones!")
	})

	buttonGrid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			if brosButton.HasFocus() {
				mod.app.SetFocus(chatButton)
			} else if chatButton.HasFocus() {
				mod.app.SetFocus(settingsButton)
			} else if settingsButton.HasFocus() {
				mod.app.SetFocus(logoutButton)
			} else if logoutButton.HasFocus() {
				mod.app.SetFocus(brosButton)
			}
		} else if event.Key() == tcell.KeyBacktab {
			if logoutButton.HasFocus() {
				mod.app.SetFocus(settingsButton)
			} else if settingsButton.HasFocus() {
				mod.app.SetFocus(chatButton)
			} else if chatButton.HasFocus() {
				mod.app.SetFocus(brosButton)
			} else if brosButton.HasFocus() {
				mod.app.SetFocus(logoutButton)
			}
		}
		return event
	})

	buttonGrid.SetRows(3, 1, 1).
		SetColumns(0, 1, 0, 1, 0, 1, 0)

	buttonGrid.AddItem(brosButton, 0, 0, 1, 1, 0, 0, true).
		AddItem(chatButton, 0, 2, 1, 1, 0, 0, true).
		AddItem(settingsButton, 0, 4, 1, 1, 0, 0, true).
		AddItem(logoutButton, 0, 6, 1, 1, 0, 0, true).
		AddItem(tvInstructions, 2, 0, 1, 7, 0, 0, false)

	grid.AddItem(logoBro, 1, 1, 1, 1, 0, 0, false).
		AddItem(logoChat, 1, 2, 1, 1, 0, 0, false).
		AddItem(buttonGrid, 2, 1, 1, 2, 0, 0, true)

	mod.pageNav.Register(HOME_MENU_PAGE, grid, true, false, func() {
		// Make a call to get the user
		ses, ok := state.Get[state.UserSession](mod.appState, state.UserSessionProp)

		if !ok {
			AlertFatal(mod.app, mod.pageNav.Pages, "home:menu:alert:err", "User Session Not Valid")
		}

		// Make sure the session is still valid
		if ses.Auth.TokenExpiration.Before(time.Now()) {
			state.Set(mod.appState, state.UserSessionProp, nil)
			// Send user to the login page
			mod.pageNav.NavigateTo(LOGIN_PAGE)
		}
	})
}

const (
	FRIENDS_LIST_PAGE_ALERT_INFO = "home:friendlist:alert:info"
	FRIENDS_LIST_PAGE_ALERT_ERR  = "home:friendlist:alert:err"
)

func (mod *HomeModule) setupFriendListPage() {
	tvHeader := tview.NewTextView().SetTextAlign(tview.AlignCenter)
	tvHeader.SetBackgroundColor(DefaultBackgroundColor)
	tvHeader.SetTextColor(tcell.NewHexColor(0xFFFFFF))
	tvHeader.SetText("Friends List")

	userFriends := make(map[uint8]chat.UserRelationship, 0)

	table := tview.NewTable().
		SetBorders(true)
	table.SetBackgroundColor(DefaultBackgroundColor)
	table.SetFixed(1, 1)
	table.SetSelectable(true, false)

	table.SetSelectedFunc(func(row int, _ int) {
		rel, ok := userFriends[uint8(row)]

		if !ok {
			return
		}

		Alert(mod.pageNav.Pages, FRIENDS_LIST_PAGE_ALERT_INFO, fmt.Sprintf("Selected User: %s", rel.Username))
	})

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 'q':
				mod.pageNav.NavigateTo(HOME_MENU_PAGE)
				userFriends = make(map[uint8]chat.UserRelationship, 0)
				table.Clear()
			case 'f':
				mod.pageNav.NavigateTo(HOME_FRIENDS_FINDER_PAGE)
				userFriends = make(map[uint8]chat.UserRelationship, 0)
				table.Clear()
			case 'p':
				Alert(mod.pageNav.Pages, FRIENDS_LIST_PAGE_ALERT_INFO, "Pending Requests Feature Not Implemented Yet")
			}
		}

		return event
	})

	tvInstructions := tview.NewTextView().SetTextAlign(tview.AlignCenter)
	tvInstructions.SetBackgroundColor(DefaultBackgroundColor)
	tvInstructions.SetTextColor(tcell.NewHexColor(0xFFFFFF))
	tvInstructions.SetText("(f) Find a new Bro - (p) View Pending - (q) Quit")

	grid := tview.NewGrid()
	grid.SetBackgroundColor(DefaultBackgroundColor)

	grid.SetRows(2, 1, 1, 0, 1, 1, 2)
	grid.SetColumns(0, 76, 0)

	grid.AddItem(tvHeader, 1, 1, 1, 1, 0, 0, false)
	grid.AddItem(table, 3, 1, 1, 1, 0, 0, true)
	grid.AddItem(tvInstructions, 5, 1, 1, 1, 0, 0, false)

	mod.pageNav.Register(HOME_FRIENDS_LIST_PAGE, grid, true, false, func() {
		userFriends = make(map[uint8]chat.UserRelationship, 0)
		table.Clear()

		table.SetCell(0, 0, tview.NewTableCell("Username").
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignCenter).
			SetExpansion(1).
			SetSelectable(false).
			SetAttributes(tcell.AttrBold|tcell.AttrUnderline))

		table.SetCell(0, 1, tview.NewTableCell("Status").
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignCenter).
			SetSelectable(false).
			SetAttributes(tcell.AttrBold|tcell.AttrUnderline))

		table.SetCell(0, 2, tview.NewTableCell("Last Active").
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignRight).
			SetSelectable(false).
			SetAttributes(tcell.AttrBold|tcell.AttrUnderline))

		session, ok := state.Get[state.UserSession](mod.appState, state.UserSessionProp)

		if !ok {
			AlertFatal(mod.app, mod.pageNav.Pages, FRIENDS_LIST_PAGE_ALERT_ERR, "Application State Error - Could not get user session.")
			return
		}

		usr, err := mod.brochatClient.GetUser(&chat.AuthInfo{
			AccessToken: session.Auth.AccessToken,
			TokenType:   "Bearer",
		}, session.Info.Id)

		if err != nil {
			AlertFatal(mod.app, mod.pageNav.Pages, FRIENDS_LIST_PAGE_ALERT_ERR, err.Error())
			return
		}

		for i, rel := range usr.Relationships {
			row := i + 1

			if rel.Type != chat.RELATIONSHIP_TYPE_FRIEND {
				continue
			}

			table.SetCell(row, 0, tview.NewTableCell(rel.Username).SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignCenter))
			if rel.IsOnline {
				table.SetCell(row, 1, tview.NewTableCell("Online").SetTextColor(tcell.ColorGreen).SetAlign(tview.AlignCenter))
			} else {
				table.SetCell(row, 1, tview.NewTableCell("Offline").SetTextColor(tcell.ColorGray).SetAlign(tview.AlignCenter))
			}

			var dateString string = rel.LastOnlineUtc.Local().Format("Jan 2, 2006")

			table.SetCell(row, 2, tview.NewTableCell(dateString).SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignRight))

			userFriends[uint8(row)] = rel
		}
	})
}

const (
	FIND_A_FRIEND_PAGE_ALERT_INFO = "home:findafriend:alert:info"
	FIND_A_FRIEND_PAGE_ALERT_ERR  = "home:findafriend:alert:err"
	FIND_A_FRIEND_PAGE_CONFIRM    = "home:findafriend:confirm"
)

func (mod *HomeModule) setupFindAFriendPage() {
	tvHeader := tview.NewTextView().SetTextAlign(tview.AlignCenter)
	tvHeader.SetBackgroundColor(DefaultBackgroundColor)
	tvHeader.SetTextColor(tcell.NewHexColor(0xFFFFFF))
	tvHeader.SetText("Find Friends")

	users := make(map[uint8]chat.UserInfo, 0)

	table := tview.NewTable().
		SetBorders(true)
	table.SetBackgroundColor(DefaultBackgroundColor)
	table.SetFixed(1, 1)
	table.SetSelectable(true, false)

	table.SetSelectedFunc(func(row int, _ int) {
		uInfo, ok := users[uint8(row)]

		if !ok {
			return
		}

		Confirm(mod.pageNav.Pages, FIND_A_FRIEND_PAGE_CONFIRM, fmt.Sprintf("Send Friend Request to %s?", uInfo.Username), func() {
			session, ok := state.Get[state.UserSession](mod.appState, state.UserSessionProp)

			if !ok {
				AlertFatal(mod.app, mod.pageNav.Pages, FIND_A_FRIEND_PAGE_ALERT_ERR, "Application State Error - Could not get user session.")
				return
			}

			err := mod.brochatClient.SendFriendRequest(&chat.AuthInfo{
				AccessToken: session.Auth.AccessToken,
				TokenType:   "Bearer",
			}, &chat.SendFriendRequestRequest{
				RequestedUserId: uInfo.ID,
			})

			if err != nil {
				if err.Error() == "friend request already exists or users are already a friend" {
					Alert(mod.pageNav.Pages, FIND_A_FRIEND_PAGE_ALERT_INFO, fmt.Sprintf("Friend Request Already Sent to %s", uInfo.Username))
					return
				} else if err.Error() == "user not found" {
					Alert(mod.pageNav.Pages, FIND_A_FRIEND_PAGE_ALERT_INFO, fmt.Sprintf("User %s Not Found", uInfo.Username))
					return
				}

				AlertFatal(mod.app, mod.pageNav.Pages, FIND_A_FRIEND_PAGE_ALERT_ERR, err.Error())
				return
			}

			Alert(mod.pageNav.Pages, FIND_A_FRIEND_PAGE_ALERT_INFO, fmt.Sprintf("Friend Request Sent to %s", uInfo.Username))
		})
	})

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 'q':
				mod.pageNav.NavigateTo(HOME_MENU_PAGE)
				users = make(map[uint8]chat.UserInfo, 0)
				table.Clear()
			}
		}

		return event
	})

	tvInstructions := tview.NewTextView().SetTextAlign(tview.AlignCenter)
	tvInstructions.SetBackgroundColor(DefaultBackgroundColor)
	tvInstructions.SetTextColor(tcell.NewHexColor(0xFFFFFF))
	tvInstructions.SetText("(s) Send Friend Request - (q) Quit")

	grid := tview.NewGrid()
	grid.SetBackgroundColor(DefaultBackgroundColor)

	grid.SetRows(2, 1, 1, 0, 1, 1, 2)
	grid.SetColumns(0, 76, 0)

	grid.AddItem(tvHeader, 1, 1, 1, 1, 0, 0, false)
	grid.AddItem(table, 3, 1, 1, 1, 0, 0, true)
	grid.AddItem(tvInstructions, 5, 1, 1, 1, 0, 0, false)

	mod.pageNav.Register(HOME_FRIENDS_FINDER_PAGE, grid, true, false, func() {
		users = make(map[uint8]chat.UserInfo, 0)
		table.Clear()

		table.SetCell(0, 0, tview.NewTableCell("Username").
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignCenter).
			SetExpansion(1).
			SetSelectable(false).
			SetAttributes(tcell.AttrBold|tcell.AttrUnderline))

		table.SetCell(0, 1, tview.NewTableCell("Last Active").
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignRight).
			SetSelectable(false).
			SetAttributes(tcell.AttrBold|tcell.AttrUnderline))

		session, ok := state.Get[state.UserSession](mod.appState, state.UserSessionProp)

		if !ok {
			AlertFatal(mod.app, mod.pageNav.Pages, FIND_A_FRIEND_PAGE_ALERT_ERR, "Application State Error - Could not get user session.")
			return
		}

		usrs, err := mod.brochatClient.GetUsers(&chat.AuthInfo{
			AccessToken: session.Auth.AccessToken,
			TokenType:   "Bearer",
		}, true, true, 1, 10, "")

		if err != nil {
			AlertFatal(mod.app, mod.pageNav.Pages, FIND_A_FRIEND_PAGE_ALERT_ERR, err.Error())
			return
		}

		for i, usr := range usrs {
			row := i + 1

			table.SetCell(row, 0, tview.NewTableCell(usr.Username).SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignCenter))
			var dateString string = usr.LastOnlineUtc.Local().Format("Jan 2, 2006")
			table.SetCell(row, 1, tview.NewTableCell(dateString).SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignRight))

			users[uint8(row)] = usr
		}
	})
}
