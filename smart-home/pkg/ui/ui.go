package ui

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/go-logr/logr"

	"github.com/loodse/godays-2020-k8s-workshop/smart-home/pkg/smarthome"
)

type UI struct {
	client  *smarthome.Client
	redraw  time.Duration
	closeCh chan struct{}
	logger  *uiLogger
	log     ui.Drawable
}

func NewUI(client *smarthome.Client, redraw time.Duration) *UI {
	return &UI{
		client:  client,
		redraw:  redraw,
		closeCh: make(chan struct{}),
		logger:  newUILogger(),
	}
}

func (u *UI) Logger() logr.Logger {
	return u.logger
}

func (u *UI) CloseCh() <-chan struct{} {
	return u.closeCh
}

func (u *UI) Run() {
	defer close(u.closeCh)

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	u.draw()

	ticker := time.NewTicker(u.redraw)
	defer ticker.Stop()

	uiEvents := ui.PollEvents()
	for {
		select {
		case <-ticker.C:
			u.draw()
			continue

		case <-u.logger.sink.updated:

		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			case "<Down>":
				u.logger.sink.log.ScrollDown()
			case "<Up>":
				u.logger.sink.log.ScrollUp()
			case "<Right>":
				u.logger.sink.log.ScrollRight()
			case "<Left>":
				u.logger.sink.log.ScrollLeft()
			}
			ui.Render(u.log)
		}
	}
}

func (u *UI) draw() {
	var elements []ui.Drawable

	shuttersTitle := widgets.NewParagraph()
	shuttersTitle.Text = "Shutters"
	shuttersTitle.SetRect(0, 0, 50, 1)
	shuttersTitle.Border = false

	elements = append(elements, shuttersTitle)

	shutters, _ := u.client.ListShutterStates()
	pos := 2
	for _, shutter := range shutters {
		g := widgets.NewGauge()
		g.Title = " " + shutter.Name + " "
		if shutter.IsMoving {
			g.Title = g.Title + "<moving> "
		}
		g.Percent = shutter.Current
		g.BarColor = ui.ColorBlue
		g.LabelStyle = ui.NewStyle(ui.ColorBlue)
		g.BorderStyle.Fg = ui.ColorWhite
		g.SetRect(0, pos, 50, pos+3)
		pos += 3

		elements = append(elements, g)
	}

	grid := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, pos, termWidth, termHeight)
	grid.Set(
		ui.NewRow(1, u.logger.sink.log),
	)
	u.log = grid
	elements = append(elements, grid)

	ui.Clear()
	ui.Render(elements...)
}

type uiLogSink struct {
	log     *widgets.List
	updated chan struct{}
	sync.Mutex
}

func newUILogSink() *uiLogSink {
	log := widgets.NewList()
	log.Title = "Logs"
	log.TextStyle = ui.NewStyle(ui.ColorWhite)
	log.SelectedRowStyle = ui.NewStyle(ui.ColorBlue)
	log.WrapText = false

	return &uiLogSink{
		log:     log,
		updated: make(chan struct{}, 100),
	}
}

func (s *uiLogSink) Log(line string) {
	s.Lock()
	defer s.Unlock()
	s.log.Rows = append(s.log.Rows, line)
	s.updated <- struct{}{}
}

var _ logr.Logger = (*uiLogger)(nil)

type uiLogger struct {
	sink   *uiLogSink
	names  []string
	values map[string]interface{}
}

func newUILogger() *uiLogger {
	return &uiLogger{
		sink: newUILogSink(),
	}
}

func (l *uiLogger) Enabled() bool {
	return true
}

func (l *uiLogger) Info(msg string, kvs ...interface{}) {
	values := addValues(l.values, kvs...)

	j, err := json.Marshal(values)
	if err != nil {
		panic(err)
	}
	l.sink.Log(fmt.Sprintf("%-15s %-20s %s", strings.Join(l.names, "."), msg, string(j)))
	l.sink.log.ScrollBottom()
}

func (l *uiLogger) Error(err error, msg string, kvs ...interface{}) {
	l.Info(msg, append(kvs, "error", err.Error())...)
}

func (l *uiLogger) V(level int) logr.InfoLogger {
	return l
}

func (l *uiLogger) WithValues(kvs ...interface{}) logr.Logger {
	return &uiLogger{
		sink:   l.sink,
		names:  l.names,
		values: addValues(l.values, kvs),
	}
}

func (l *uiLogger) WithName(name string) logr.Logger {
	return &uiLogger{
		sink:   l.sink,
		names:  append(l.names, name),
		values: l.values,
	}
}

func addValues(base map[string]interface{}, kvs ...interface{}) map[string]interface{} {
	values := map[string]interface{}{}
	// add existing k/v pairs
	for k := range base {
		values[k] = base[k]
	}
	// add new k/v pairs
	for i := 0; i < len(kvs); i += 2 {
		if i+1 >= len(kvs) {
			return values
		}
		values[fmt.Sprint(kvs[i])] = kvs[i+1]
	}
	return values
}
