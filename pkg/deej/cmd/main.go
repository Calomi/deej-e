package main

import (
	"errors"
	"flag"
	"fmt"

	"github.com/omriharel/deej/pkg/deej"

	ole "github.com/go-ole/go-ole"
	//wca "github.com/moutend/go-wca"
)

var (
	gitCommit  string
	versionTag string
	buildType  string

	verbose bool
)

func init() {
	flag.BoolVar(&verbose, "verbose", false, "show verbose logs (useful for debugging serial)")
	flag.BoolVar(&verbose, "v", false, "shorthand for --verbose")
	flag.Parse()
}

func main() {

	// first we need a logger
	logger, err := deej.NewLogger(buildType)
	if err != nil {
		panic(fmt.Sprintf("Failed to create logger: %v", err))
	}

	named := logger.Named("main")
	named.Debug("Created logger")

	named.Infow("Version info",
		"gitCommit", gitCommit,
		"versionTag", versionTag,
		"buildType", buildType)

	// provide a fair warning if the user's running in verbose mode
	if verbose {
		named.Debug("Verbose flag provided, all log messages will be shown")
	}

	// Initialize COM first to avoid issues with session finder initialization in session_finder_windows.go#L166
	// it could be above the logger creation
	if err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
		// Ignore E_FALSE as it means it's already initialized
		const eFalse = 1
		oleError := &ole.OleError{}
		fmt.Printf("OleError: %+v\n", oleError)

		if errors.As(err, &oleError) {
			if oleError.Code() == eFalse {
				fmt.Println("CoInitializeEx in main failed with E_FALSE due to redundant invocation")
			} else {
				fmt.Println("Error al llamar CoInitializeEx en main:",
					"isOleError", true,
					"error", err,
					"oleError", oleError)

				fmt.Errorf("call CoInitializeEx: %w", err)
			}
		} else {
			panic(fmt.Sprintf("Failed to call CoInitializeEx in main: %v", err))
		}
	}
	defer ole.CoUninitialize()

	// create the deej instance
	d, err := deej.NewDeej(logger, verbose)
	if err != nil {
		named.Fatalw("Failed to create deej object", "error", err)
	}

	// if injected by build process, set version info to show up in the tray
	if buildType != "" && (versionTag != "" || gitCommit != "") {
		identifier := gitCommit
		if versionTag != "" {
			identifier = versionTag
		}

		versionString := fmt.Sprintf("Version %s-%s", buildType, identifier)
		d.SetVersion(versionString)
	}

	// onwards, to glory
	if err = d.Initialize(); err != nil {
		named.Fatalw("Failed to initialize deej", "error", err)
	}
}
