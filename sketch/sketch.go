package sketch

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/PuerkitoBio/goquery"
	version "github.com/hashicorp/go-version"
)

// Release describes a specific version of Sketch.
type Release struct {
	Version     *version.Version
	ReleaseDate time.Time
	DownloadURL string
}

// UnmarshalJSON implements the Unmarshaler inferface.
func (v *Release) UnmarshalJSON(b []byte) error {
	var m struct {
		Release     string
		ReleaseDate time.Time
		DownloadURL string
	}
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}

	release, err := version.NewVersion(m.Release)
	if err != nil {
		return err
	}

	v.Version = release
	v.ReleaseDate = m.ReleaseDate
	v.DownloadURL = m.DownloadURL

	return nil
}

// MarshalJSON implements the Marshaler inferface.
func (v Release) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{
		"Release":     v.Version.String(),
		"ReleaseDate": v.ReleaseDate.String(),
		"DownloadURL": v.DownloadURL,
	})
}

// GetVersions will return all the versions of Sketch.
func GetVersions() ([]Release, error) {
	doc, err := goquery.NewDocument("https://www.sketchapp.com/updates/")
	if err != nil {
		return nil, err
	}

	versions := make([]Release, 0)
	doc.Find(".update-version").Each(func(i int, s *goquery.Selection) {
		release, found := s.Attr("data-release")
		if !found {
			return
		}

		releaseVersion, err := version.NewVersion(release)
		if err != nil {
			return
		}

		releaseDateText, found := s.Attr("data-release-date")
		if !found {
			return
		}

		releaseDate, err := time.Parse("02-01-2006", releaseDateText)
		if err != nil {
			return
		}

		downloadURL, found := s.Find("a.update-download").Attr("href")
		if !found {
			return
		}

		versions = append(versions, Release{
			Version:     releaseVersion,
			ReleaseDate: releaseDate,
			DownloadURL: downloadURL,
		})
	})

	return versions, nil
}

func CheckLicense(license string) (*time.Time, error) {
	val := url.Values{}
	val.Add("license-key", license)
	val.Add("number_of_seats", "0")

	req, err := http.NewRequest("POST", "https://api.sketchapp.com/1/license/renew/", bytes.NewBufferString(val.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", "sketchversion/1.0.0")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var payload struct {
		Data struct {
			CurrentUpdateExpiration int64 `json:"current_update_expiration"`
		} `json:"data"`
	}

	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return nil, err
	}

	expirationDate := time.Unix(payload.Data.CurrentUpdateExpiration, 0)

	return &expirationDate, nil
}

func FindLatestReleaseForLicense(expiry time.Time, releases []Release) (*Release, error) {
	sort.SliceStable(releases, func(i, j int) bool {
		return releases[i].Version.LessThan(releases[j].Version)
	})

	sort.SliceStable(releases, func(i, j int) bool {
		return releases[i].ReleaseDate.After(releases[j].ReleaseDate)
	})

	for _, version := range releases {
		if version.ReleaseDate.Before(expiry) {
			return &version, nil
		}
	}

	return nil, errors.New("cannot find any valid version")
}

// Download will fetch the sketch zip file for the provided version.
func Download(version Release) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dest := filepath.Join(cwd, filepath.Base(version.DownloadURL))

	res, err := http.Get(version.DownloadURL)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	file, err := os.Create(dest)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := io.Copy(file, res.Body); err != nil {
		return "", err
	}

	return dest, nil
}
