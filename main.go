package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/wyattjoh/sketchversion/sketch"
)

func main() {
	download := flag.Bool("download", false, "download the zip automatically")

	flag.Parse()

	license := flag.Arg(0)

	current, err := sketch.CheckLicense(license)
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't get the current license: %v\n", err)
		os.Exit(1)
	}

	versions, err := sketch.GetVersions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't get the versions: %v\n", err)
		os.Exit(1)
	}

	version, err := sketch.FindLatestReleaseForLicense(*current, versions)
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't get the version: %v\n", err)
		os.Exit(1)
	}

	if *download {
		fmt.Printf("Matched to version %s, downloading\n", version.Version)

		path, err := sketch.Download(*version)
		if err != nil {
			fmt.Fprintf(os.Stderr, "can't download the version: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Downloaded version %s to %s\n", version.Version, path)
	} else {
		fmt.Printf("Your most recent version is %s.\n\n\tDownload: %s\n\n", version.Version, version.DownloadURL)
	}
}
