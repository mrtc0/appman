package application

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mitchellh/go-ps"
	"github.com/mrtc0/appman/application/config"
	"github.com/rivo/tview"
)

const (
	appmanExitConfirmMessage = "Do you want to exit appman?"
)

type TuiApplicationManager struct {
	tuiApp *tview.Application

	Layout *tview.Grid
	// Page for application list
	ApplicationPages *tview.Pages
	// Page for application log
	LogPages *tview.Pages
	// This table displays a list of applications.
	// One application is displayed per row.
	ApplicationTable *tview.Table

	// writer for writing application logs
	logWriter tview.TextViewWriter
}

func (m *TuiApplicationManager) Run() error {
	return m.tuiApp.Run()
}

func NewTuiApplicationManager(applicationConfig []config.ApplicationConfig) *TuiApplicationManager {
	tviewApp := tview.NewApplication()
	appPages := tview.NewPages()
	logPages := tview.NewPages()

	m := &TuiApplicationManager{
		tuiApp:           tviewApp,
		Layout:           NewLayout(),
		ApplicationPages: appPages,
		LogPages:         logPages,
	}

	m.AddLogView()

	var applications []Application
	for _, conf := range applicationConfig {
		logger := NewLogger(conf.Name, m.logWriter)
		applications = append(applications, NewApplication(conf, logger))
	}

	m.AddApplicationView(applications)
	m.SetupCleanupFunction()

	m.tuiApp.SetRoot(m.Layout, true).SetFocus(m.ApplicationPages)

	return m
}

func ApplicationTableHeader() string {
	return fmt.Sprintf("%-30s | %-10s | %-6s | %-6s | %-30s", "Name", "Status", "PID", "Port", "URL")
}

func (m *TuiApplicationManager) Refresh(interval time.Duration) {
	for {
		time.Sleep(interval)
		m.tuiApp.QueueUpdateDraw(func() {
			for i := 1; i < m.ApplicationTable.GetRowCount(); i++ {
				cell := m.ApplicationTable.GetCell(i, 0)

				app := cell.Reference.(Application)

				if app.status == Stopped {
					continue
				}

				p, _ := ps.FindProcess(app.Pid)
				if p == nil {
					app.SetStatus(Error)
					newCell := tview.NewTableCell(app.DisplayTableCellText())
					newCell.SetTextColor(tcell.ColorRed)
					newCell.Reference = app

					m.ApplicationTable.SetCell(i, 0, newCell)
				}
			}
		})
	}
}

// PopupApplicationActionModal is displays a modal for operating the application
func (m *TuiApplicationManager) PopupApplicationActionModal(row, column int) {
	cell := m.ApplicationTable.GetCell(row, column)
	app := cell.Reference.(Application)

	modal := tview.NewModal().SetText(app.LaunchMessage()).AddButtons(app.ActionLabels()).SetDoneFunc(func(_ int, buttonLabel string) {
		switch buttonLabel {
		case "Shutdown":
			app = *app.Stop()
			newCell := tview.NewTableCell(app.DisplayTableCellText())
			newCell.Reference = app

			m.ApplicationTable.SetCell(row, column, newCell)
		case "Launch":
			app = *app.Start()
			newCell := tview.NewTableCell(app.DisplayTableCellText())
			newCell.Reference = app
			newCell.SetTextColor(tcell.ColorGreen)

			m.ApplicationTable.SetCell(row, column, newCell)
		case "Cancel":
		}
		m.ApplicationPages.RemovePage("action")
	})

	m.ApplicationPages.AddPage("action", modal, true, true)
}

func (m *TuiApplicationManager) PopupExitAppmanConfirmModal() {
	actions := []string{"Cancel", "Exit"}
	modal := tview.NewModal().SetText(appmanExitConfirmMessage).AddButtons(actions).SetDoneFunc(func(_ int, buttonLabel string) {
		switch buttonLabel {
		case "Exit":
			m.tuiApp.Stop()
		}

		m.ApplicationPages.RemovePage("action")
	})

	m.ApplicationPages.AddPage("action", modal, true, true)
}

func (m *TuiApplicationManager) AddApplicationView(applications []Application) *TuiApplicationManager {
	table := tview.NewTable().SetBorders(true)

	table.SetCell(0, 0, tview.NewTableCell(ApplicationTableHeader()).SetAlign(tview.AlignLeft).SetExpansion(1).SetSelectable(false))
	for i, app := range applications {
		cell := tview.NewTableCell(app.DisplayTableCellText()).SetAlign(tview.AlignLeft).SetExpansion(1)
		cell.Reference = app

		table.SetCell(i+1, 0, cell)
	}

	table.Select(0, 0).SetFixed(1, 1).SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			m.PopupExitAppmanConfirmModal()
		}
		if key == tcell.KeyEnter {
			table.SetSelectable(true, true)
		}
	}).SetSelectedFunc(func(row int, column int) {
		m.PopupApplicationActionModal(row, column)
		table.SetSelectable(false, false)
	})

	m.ApplicationTable = table
	m.ApplicationPages.AddPage("application", m.ApplicationTable, true, true)
	m.Layout.AddItem(m.ApplicationPages, 1, 0, 1, 1, 0, 0, true)
	return m
}

func (m *TuiApplicationManager) AddLogView() *TuiApplicationManager {
	logs := tview.NewTextView().SetDynamicColors(true).SetChangedFunc(func() {
		m.tuiApp.Draw()
	})

	logs.SetBorder(true).SetTitle("Logs")
	logs.SetDoneFunc(func(key tcell.Key) {
		logs.ScrollToEnd()
	})

	m.LogPages.AddPage("log", logs, true, true)
	m.Layout.AddItem(m.LogPages, 2, 0, 1, 1, 0, 0, false)

	m.logWriter = logs.BatchWriter()
	defer m.logWriter.Close()

	return m
}

func (m *TuiApplicationManager) SetupCleanupFunction() {
	m.tuiApp.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTAB:
			if m.LogPages.HasFocus() {
				m.tuiApp.SetFocus(m.ApplicationPages)
			} else {
				m.tuiApp.SetFocus(m.LogPages)
			}

		case tcell.KeyCtrlC:
			for i := 1; i < m.ApplicationTable.GetRowCount(); i++ {
				cell := m.ApplicationTable.GetCell(i, 0)

				app := cell.Reference.(Application)
				if app.status == Running {
					app.Stop()
				}
			}

			m.tuiApp.Stop()
			return nil
		}

		return event
	})
}

func NewLayout() *tview.Grid {
	grid := tview.NewGrid().SetRows(2).SetBorders(false)
	return grid
}
