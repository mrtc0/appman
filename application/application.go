package application

import (
	"fmt"
	"hash/fnv"
	"os"
	"os/exec"
	"syscall"

	"github.com/fatih/color"
	"github.com/mrtc0/appman/application/config"
	"github.com/rivo/tview"
)

type ApplicationStatus int

const (
	Starting ApplicationStatus = iota
	Running
	Stopped
	Error
)

var colorList = []*color.Color{
	color.New(color.FgHiCyan),
	color.New(color.FgHiGreen),
	color.New(color.FgHiMagenta),
	color.New(color.FgHiYellow),
	color.New(color.FgHiBlue),
	color.New(color.FgHiRed),
}

func (a ApplicationStatus) String() string {
	switch a {
	case Starting:
		return "Starting"
	case Running:
		return "Running"
	case Stopped:
		return "Stopped"
	case Error:
		return "Error"
	default:
		return "Unknown"
	}
}

type Application struct {
	Name string
	Port int

	Pid    int
	status ApplicationStatus

	Path         string
	StartCommand []string
	StopCommand  []string
	Env          []string
	Logger       *Logger

	URL    string
	Branch string
}

type LogWriter interface {
	Write(p []byte) (n int, err error)
}

type Logger struct {
	applicationName string
	writer          LogWriter
	color           *color.Color
}

func determineColor(applicationName string) *color.Color {
	hash := fnv.New32()
	_, _ = hash.Write([]byte(applicationName))
	idx := hash.Sum32() % uint32(len(colorList))

	return colorList[idx]
}

func NewLogger(applicationName string, bw LogWriter) *Logger {
	lw := &Logger{
		applicationName: applicationName,
		writer:          bw,
		color:           determineColor(applicationName),
	}
	return lw
}

func (l Logger) Write(p []byte) (n int, err error) {
	f := l.color.SprintFunc()

	fmt.Fprint(
		l.writer,
		tview.TranslateANSI(
			fmt.Sprintf("%s\t%s", f(l.applicationName), string(p)),
		),
	)

	return len(p), nil
}

func NewApplication(conf config.ApplicationConfig, logger *Logger) Application {
	return Application{
		Name:         conf.Name,
		status:       Stopped,
		Path:         conf.Path,
		StartCommand: conf.StartCommand,
		StopCommand:  conf.StopCommand,
		Env:          conf.Env,
		URL:          conf.URL,
		Port:         conf.Port,
		Branch:       "TODO",
		Logger:       logger,
	}
}

func (app *Application) DisplayTableCellText() string {
	return fmt.Sprintf("%-30s | %-10s | %-6d | %-6d | %-30s", app.Name, app.status, app.Pid, app.Port, app.URL)
}

func (app Application) LaunchMessage() string {
	return fmt.Sprintf("Do you want launch the application: %s", app.Name)
}

func (app Application) ActionLabels() []string {
	switch app.status {
	case Running:
		return []string{"Shutdown", "Cancel"}
	case Stopped:
		return []string{"Launch", "Cancel"}
	case Error:
		return []string{"Launch", "Cancel"}
	default:
		return []string{"Cancel"}
	}
}

func (app *Application) SetStatus(status ApplicationStatus) {
	app.status = status
}

// Start is start application
func (app *Application) Start() *Application {
	app.SetStatus(Starting)

	err := app.start()
	if err != nil {
		return app
	}

	app.SetStatus(Running)
	return app
}

// Stop is sends a signal to the process to stop it.
func (app *Application) Stop() *Application {
	if app.StopCommand != nil {
		err := app.stop()
		if err != nil {
			return app
		}
	} else {
		err := kill(app.Pid)
		if err != nil {
			return app
		}
	}

	app.SetStatus(Stopped)

	return app
}

func (app *Application) start() error {
	cmd := exec.Command(app.StartCommand[0], app.StartCommand[1:]...)
	cmd.Dir = app.Path
	cmd.Env = append(os.Environ(), app.Env...)
	cmd.Stdout = app.Logger
	cmd.Stderr = app.Logger

	err := cmd.Start()
	if err != nil {
		return err
	}

	app.Pid = cmd.Process.Pid

	go func() {
		cmd.Wait()
	}()

	return nil
}

func (app *Application) stop() error {
	cmd := exec.Command(app.StopCommand[0], app.StopCommand[1:]...)
	cmd.Dir = app.Path
	cmd.Env = append(os.Environ(), app.Env...)
	cmd.Stdout = app.Logger

	app.Pid = 0

	return cmd.Start()
}

func kill(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	return process.Signal(syscall.SIGTERM)
}
