package ui

import (
	"context"
	"fmt"
	"log"

	"github.com/dmars8047/brolib/chat"
	"github.com/dmars8047/broterm/internal/state"
	"github.com/dmars8047/broterm/internal/theme"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const ACCEPT_FRIEND_REQUEST_PAGE PageSlug = "accept_friend_request"

// AcceptFriendRequestPage is the page for accepting friend requests
type AcceptFriendRequestPage struct {
	brochatClient       *chat.BroChatClient
	userPendingRequests map[uint8]chat.UserRelationship
	table               *tview.Table
	feedClient          *state.FeedClient
	currentThemeCode    string
}

// NewAcceptFriendRequestPage creates a new accept friend request page
func NewAcceptFriendRequestPage(brochatClient *chat.BroChatClient, feedClient *state.FeedClient) *AcceptFriendRequestPage {
	return &AcceptFriendRequestPage{
		brochatClient:       brochatClient,
		feedClient:          feedClient,
		userPendingRequests: make(map[uint8]chat.UserRelationship, 0),
		table:               tview.NewTable(),
		currentThemeCode:    "NOT_SET",
	}
}

// Setup sets up the accept friend request page and registers it with the page navigator
func (page *AcceptFriendRequestPage) Setup(app *tview.Application, appContext *state.ApplicationContext, nav *PageNavigator) {

	theme := appContext.GetTheme()

	tvHeader := tview.NewTextView().SetTextAlign(tview.AlignCenter)
	tvHeader.SetText("Pending Friend Requests")

	page.table.SetBorders(true)
	page.table.SetFixed(1, 1)
	page.table.SetSelectable(true, false)

	page.table.SetSelectedFunc(func(row int, _ int) {
		selectedUser, ok := page.userPendingRequests[uint8(row)]

		if !ok {
			return
		}

		accessToken, ok := appContext.GetAccessToken()

		if !ok {
			log.Printf("Valid user authentication information not found. Redirecting to login page.")
			nav.NavigateTo(LOGIN_PAGE, nil)
			return
		}

		nav.Confirm(FIND_A_FRIEND_PAGE_CONFIRM, fmt.Sprintf("Accept Friend Request from %s?", selectedUser.Username), func() {
			result := page.brochatClient.AcceptFriendRequest(accessToken, chat.AcceptFriendRequestRequest{
				InitiatingUserId: selectedUser.UserId,
			})

			err := result.Err()

			if err != nil {
				if len(result.ErrorDetails) > 0 {
					nav.Alert(FIND_A_FRIEND_PAGE_ALERT_INFO, result.ErrorDetails[0])
					return
				}

				if result.ResponseCode == chat.BROCHAT_RESPONSE_CODE_FORBIDDEN_ERROR {
					nav.Alert(FIND_A_FRIEND_PAGE_ALERT_INFO, FORBIDDEN_OPERATION_ERROR_MESSAGE)
					return
				}

				nav.AlertFatal(app, FIND_A_FRIEND_PAGE_ALERT_ERR, err.Error())
				return
			}

			page.table.RemoveRow(row)
			nav.Alert(FIND_A_FRIEND_PAGE_ALERT_INFO, fmt.Sprintf("Accepted Friend Request from %s", selectedUser.Username))
		})
	})

	page.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			nav.NavigateTo(FRIENDS_LIST_PAGE, nil)
			page.userPendingRequests = make(map[uint8]chat.UserRelationship, 0)
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
	tvInstructions.SetText("(esc) Quit")

	grid := tview.NewGrid()
	grid.SetBackgroundColor(theme.BackgroundColor)

	grid.SetRows(2, 1, 1, 0, 1, 1, 2)
	grid.SetColumns(0, 76, 0)

	grid.AddItem(tvHeader, 1, 1, 1, 1, 0, 0, false)
	grid.AddItem(page.table, 3, 1, 1, 1, 0, 0, true)
	grid.AddItem(tvInstructions, 5, 1, 1, 1, 0, 0, false)

	var pageContext context.Context
	var cancel context.CancelFunc

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

	nav.Register(ACCEPT_FRIEND_REQUEST_PAGE, grid, true, false,
		func(param interface{}) {
			pageContext, cancel = appContext.GenerateUserSessionBoundContextWithCancel()
			applyTheme()
			page.onPageLoad(app, appContext, pageContext, page.feedClient)
		},
		func() {
			cancel()
			page.onPageClose()
		})
}

// onPageLoad is called when the page is navigated to
func (page *AcceptFriendRequestPage) onPageLoad(app *tview.Application, appContext *state.ApplicationContext,
	pageContext context.Context, feedClient *state.FeedClient) {

	page.populateTable(appContext.GetBrochatUser(), appContext.GetTheme())

	go func() {
		subId, userProfileUpdatesChannel := feedClient.SubscribeToUserProfileUpdates()

		defer feedClient.UnsubscribeFromUserProfileUpdates(subId)

		for {
			select {
			case <-pageContext.Done():
				return
			case updateCode := <-userProfileUpdatesChannel:
				if updateCode == chat.USER_PROFILE_UPDATE_REASON_RELATIONSHIP_UPDATE {
					page.table.Clear()
					app.QueueUpdateDraw(func() {
						page.populateTable(appContext.GetBrochatUser(), appContext.GetTheme())
					})
				}
			}
		}
	}()
}

// onPageClose is called when the page is navigated away from
func (page *AcceptFriendRequestPage) onPageClose() {
	page.userPendingRequests = make(map[uint8]chat.UserRelationship)
	page.table.Clear()
}

// populateTable populates the users from the brochat user's relationships into the table
func (page *AcceptFriendRequestPage) populateTable(brochatUser chat.User, thm theme.Theme) {
	page.table.SetCell(0, 0, tview.NewTableCell("Username").
		SetTextColor(thm.ForgroundColor).
		SetAlign(tview.AlignCenter).
		SetExpansion(1).
		SetSelectable(false).
		SetAttributes(tcell.AttrBold|tcell.AttrUnderline))

	page.table.SetCell(0, 1, tview.NewTableCell("Last Active").
		SetTextColor(thm.ForgroundColor).
		SetAlign(tview.AlignRight).
		SetSelectable(false).
		SetAttributes(tcell.AttrBold|tcell.AttrUnderline))

	row := 1

	for _, rel := range brochatUser.Relationships {
		if rel.Type&chat.RELATIONSHIP_TYPE_FRIEND_REQUEST_RECIEVED != 0 {
			page.table.SetCell(row, 0, tview.NewTableCell(rel.Username).SetTextColor(thm.ForgroundColor).SetAlign(tview.AlignCenter))
			var dateString string = rel.LastOnlineUtc.Local().Format("Jan 2, 2006")
			page.table.SetCell(row, 1, tview.NewTableCell(dateString).SetTextColor(thm.ForgroundColor).SetAlign(tview.AlignRight))

			page.userPendingRequests[uint8(row)] = rel
			row++
		}
	}
}
