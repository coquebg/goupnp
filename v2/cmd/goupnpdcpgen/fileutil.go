package main

import (
	"archive/zip"
	"encoding/xml"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"

	"github.com/huin/goupnp/v2/errkind"
)

func acquireFile(specFilename string, xmlSpecURL string) error {
	if f, err := os.Open(specFilename); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		f.Close()
		return nil
	}

	resp, err := http.Get(xmlSpecURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errkind.NewUnexpectedHTTPStatus(resp.StatusCode, resp.Status)
	}

	tmpFilename := specFilename + ".download"
	w, err := os.Create(tmpFilename)
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return err
	}

	return os.Rename(tmpFilename, specFilename)
}

func globFiles(pattern string, archive *zip.ReadCloser) []*zip.File {
	var files []*zip.File
	for _, f := range archive.File {
		if matched, err := path.Match(pattern, f.Name); err != nil {
			// This shouldn't happen - all patterns are hard-coded, errors in
			// them are a programming error.
			panic(err)
		} else if matched {
			files = append(files, f)
		}
	}
	return files
}

func unmarshalXMLFile(file *zip.File, data interface{}) error {
	r, err := file.Open()
	if err != nil {
		return err
	}
	decoder := xml.NewDecoder(r)
	defer r.Close()
	return decoder.Decode(data)
}

var scpdFilenameRe = regexp.MustCompile(`.*/([a-zA-Z0-9]+)([0-9]+)\.xml`)

func urnPartsFromSCPDFilename(filename string) (*urnParts, error) {
	parts := scpdFilenameRe.FindStringSubmatch(filename)
	if len(parts) != 3 {
		return nil, errkind.New(
			errkind.InvalidArgument,
			"SCPD filename %q does not have expected number of parts",
			filename,
		)
	}
	name, version := parts[1], parts[2]
	return &urnParts{
		URN:     serviceURNPrefix + name + ":" + version,
		Name:    name,
		Version: version,
	}, nil
}