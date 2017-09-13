package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/wyattjoh/sketchversion/sketch"
)

func GetSketchRelease(license string) (*sketch.Release, error) {
	current, err := sketch.CheckLicense(license)
	if err != nil {
		return nil, err
	}

	versions, err := sketch.GetVersions()
	if err != nil {
		return nil, err
	}

	release, err := sketch.FindLatestReleaseForLicense(*current, versions)
	if err != nil {
		return nil, err
	}

	return release, nil
}

func main() {
	download := flag.Bool("download", false, "download the zip automatically")

	flag.Parse()

	license := flag.Arg(0)

	release, err := GetSketchRelease(license)
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't get the release: %v\n", err)
		os.Exit(1)
	}

	if *download {
		fmt.Printf("Matched to version %s, downloading\n", release.VersionString)

		path, err := sketch.Download(*release)
		if err != nil {
			fmt.Fprintf(os.Stderr, "can't download the version: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Downloaded version %s to %s\n", release.VersionString, path)
	} else {
		fmt.Printf("Your most recent version is %s.\n\n\tDownload: %s\n\n", release.VersionString, release.DownloadURL)
	}
}
