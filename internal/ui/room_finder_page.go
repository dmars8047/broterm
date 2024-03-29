package ui

import (
	"fmt"
	"log"

	"github.com/dmars8047/brolib/chat"
	"github.com/dmars8047/broterm/internal/state"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const ROOM_FINDER_PAGE PageSlug = "room_finder"

const (
	ROOM_FINDER_PAGE_ALERT_INFO = "home:roomfinder:alert:info"
	ROOM_FINDER_PAGE_ALERT_ERR  = "home:roomfinder:alert:err"
	ROOM_FINDER_PAGE_CONFIRM    = "home:roomfinder:confirm"
)

// RoomFinderPage is the room finder page
type RoomFinderPage struct {
	brochatClient    *chat.BroChatClient
	table            *tview.Table
	publicRooms      map[int]chat.Room
	currentThemeCode string
}

// NewRoomFinderPage creates a new room finder page
func NewRoomFinderPage(brochatClient *chat.BroChatClient) *RoomFinderPage {
	return &RoomFinderPage{
		brochatClient:    brochatClient,
		table:            tview.NewTable(),
		publicRooms:      make(map[int]chat.Room, 0),
		currentThemeCode: "NOT_SET",
	}
}

// Setup sets up the room finder page and registers it with the page navigator
func (page *RoomFinderPage) Setup(app *tview.Application, appContext *state.ApplicationContext, nav *PageNavigator) {
	tvHeader := tview.NewTextView().SetTextAlign(tview.AlignCenter)
	tvHeader.SetText("Find Rooms")

	page.table.SetBorders(true)
	page.table.SetFixed(1, 1)
	page.table.SetSelectable(true, false)

	page.table.SetSelectedFunc(func(row int, _ int) {
		room, ok := page.publicRooms[row]

		if !ok {
			return
		}

		accessToken, ok := appContext.GetAccessToken()

		if !ok {
			log.Printf("Valid user authentication information not found. Redirecting to login page.")
			nav.NavigateTo(LOGIN_PAGE, nil)
			return
		}

		nav.Confirm(ROOM_FINDER_PAGE_CONFIRM, fmt.Sprintf("Join %s?", room.Name), func() {
			joinRoomResult := page.brochatClient.JoinRoom(accessToken, room.Id)

			joinRoomErr := joinRoomResult.Err()

			if joinRoomErr != nil {
				if len(joinRoomResult.ErrorDetails) > 0 {
					nav.Alert(ROOM_FINDER_PAGE_ALERT_INFO, joinRoomResult.ErrorDetails[0])
					return
				}

				if joinRoomResult.ResponseCode == chat.BROCHAT_RESPONSE_CODE_FORBIDDEN_ERROR {
					nav.Alert(ROOM_FINDER_PAGE_ALERT_INFO, FORBIDDEN_OPERATION_ERROR_MESSAGE)
					return
				}

				nav.AlertFatal(app, ROOM_FINDER_PAGE_ALERT_ERR, fmt.Sprintf("An error occurred while joining room: %s", joinRoomErr.Error()))
				return
			}

			nav.AlertWithDoneFunc(ROOM_FINDER_PAGE_ALERT_INFO, fmt.Sprintf("You have successfuly joined the room '%s'.", room.Name), func(buttonIndex int, buttonLabel string) {
				nav.Pages.HidePage(ROOM_FINDER_PAGE_ALERT_INFO).RemovePage(ROOM_FINDER_PAGE_ALERT_INFO)
				nav.NavigateTo(CHAT_PAGE, ChatPageParameters{
					channel_id: room.ChannelId,
					title:      room.Name,
					returnPage: ROOM_LIST_PAGE,
				})
			})
		})
	})

	page.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			nav.NavigateTo(ROOM_LIST_PAGE, nil)
			page.publicRooms = make(map[int]chat.Room, 0)
			page.table.Clear()
		} else if event.Key() == tcell.KeyTab {
			// Change the selected row to the next row
			row, _ := page.table.GetSelection()
			if row+1 >= page.table.GetRowCount() {
				row = 1
			} else {
				row++
			}

			page.table.Select(row, 0)
		} else if event.Key() == tcell.KeyBacktab {
			// Change the selected row to the previous row
			row, _ := page.table.GetSelection()

			if row-1 < 1 {
				row = page.table.GetRowCount() - 1
			} else {
				row--
			}

			page.table.Select(row, 0)
		}

		return event
	})

	tvInstructions := tview.NewTextView().SetTextAlign(tview.AlignCenter)
	tvInstructions.SetText("(enter) Join room - (esc) Quit")

	grid := tview.NewGrid()

	grid.SetRows(2, 1, 1, 0, 1, 1, 2)
	grid.SetColumns(0, 76, 0)

	grid.AddItem(tvHeader, 1, 1, 1, 1, 0, 0, false)
	grid.AddItem(page.table, 3, 1, 1, 1, 0, 0, true)
	grid.AddItem(tvInstructions, 5, 1, 1, 1, 0, 0, false)

	applyTheme := func() {
		theme := appContext.GetTheme()

		if page.currentThemeCode != theme.Code {
			page.currentThemeCode = theme.Code
			grid.SetBackgroundColor(theme.BackgroundColor)
			page.table.SetBordersColor(theme.BorderColor)
			page.table.SetBorderColor(theme.BorderColor)
			page.table.SetTitleColor(theme.TitleColor)
			page.table.SetBackgroundColor(theme.BackgroundColor)
			page.table.SetSelectedStyle(theme.DropdownListSelectedStyle)
			tvHeader.SetBackgroundColor(theme.BackgroundColor)
			tvHeader.SetTextColor(theme.TitleColor)
			tvInstructions.SetBackgroundColor(theme.BackgroundColor)
			tvInstructions.SetTextColor(theme.InfoColor)
		}
	}

	applyTheme()

	nav.Register(ROOM_FINDER_PAGE, grid, true, false,
		func(_ interface{}) {
			applyTheme()
			page.onPageLoad(appContext, nav)
		},
		func() {
			page.onPageClose()
		})
}

// onPageLoad is called when the room finder page is navigated to
func (page *RoomFinderPage) onPageLoad(appContext *state.ApplicationContext, nav *PageNavigator) {
	accessToken, ok := appContext.GetAccessToken()

	if !ok {
		log.Printf("Valid user authentication information not found. Redirecting to login page.")
		nav.NavigateTo(LOGIN_PAGE, nil)
		return
	}

	thm := appContext.GetTheme()

	page.table.SetCell(0, 0, tview.NewTableCell("Name").
		SetTextColor(thm.ForgroundColor).
		SetAlign(tview.AlignCenter).
		SetExpansion(1).
		SetSelectable(false).
		SetAttributes(tcell.AttrBold|tcell.AttrUnderline))

	page.table.SetCell(0, 1, tview.NewTableCell("Owner").
		SetTextColor(thm.ForgroundColor).
		SetAlign(tview.AlignCenter).
		SetSelectable(false).
		SetAttributes(tcell.AttrBold|tcell.AttrUnderline))

	getRoomsResult := page.brochatClient.GetRooms(accessToken)

	err := getRoomsResult.Err()

	if err != nil {
		nav.Alert(ROOM_FINDER_PAGE_ALERT_ERR, fmt.Sprintf("An error occurred while retrieving public rooms: %s", err.Error()))
		return
	}

	rooms := getRoomsResult.Content

	for i, rel := range rooms {
		row := i + 1

		page.table.SetCell(row, 0, tview.NewTableCell(rel.Name).SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignCenter))
		page.table.SetCell(row, 1, tview.NewTableCell(rel.Owner.Username).SetTextColor(tcell.ColorGreen).SetAlign(tview.AlignCenter))

		page.publicRooms[row] = rel
	}
}

// onPageClose is called when the room finder page is navigated away from
func (page *RoomFinderPage) onPageClose() {
	page.publicRooms = make(map[int]chat.Room, 0)
	page.table.Clear()
}
