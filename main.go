package main

import (
	"cmp"
	_ "embed"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"
)

var verbosity = false
var imgMode = false
var quiet = false

var ConsideredImage = map[string]struct{}{
	".jpg":  {},
	".jpeg": {},
	".png":  {},
	".gif":  {},
	".webp": {},
	".svg":  {},
}

var ListingIcon = map[string]string{
	"default":   "📄",
	"directory": "🗂️",
	".jpg":      "🖼️",
	".jpeg":     "🖼️",
	".png":      "🖼️",
	".gif":      "🖼️",
	".webp":     "🖼️",
	".svg":      "🖼️",
	".bmp":      "🖼️",
	".apng":     "🖼️",
	".exe":      "⚠️",
	".sh":       "⚠️",
	".pdf":      "📑",
	".lnk":      "🔗",
	".html":     "📰",
	".url":      "🌏",
	".css":      "🎨",
	".js":       "🤖",
	".go":       "🧊",
	".java":     "☕",
	".class":    "🧻",
	// OH GODS, THERE ARE LITERALLY BILLIONS OF THEM
}

type ListingEntry struct {
	Icon string
	Name string
}

//go:embed gallery.html
var rawGalleryTemplate string

//go:embed listing.html
var rawListingTemplate string

func message(format string, args ...any) {
	if quiet {
		return
	}
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}
func verbose(format string, args ...any) {
	if !verbosity {
		return
	}
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}
func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func MustGetEntries(path string) []fs.DirEntry {
	entries, err := os.ReadDir(path)
	if err != nil {
		fatal("Could not get current directory entries: %s", err.Error())
	}
	return entries
}

func gallery(path string) {
	galleryTemplate := template.Must(template.New("gallery").Parse(rawGalleryTemplate))
	entries := MustGetEntries(path)
	images := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if _, ok := ConsideredImage[ext]; ok {
			images = append(images, entry.Name())
			verbose("%s added", entry.Name())
		} else {
			verbose("%s recjected, %q is not considered an image", entry.Name(), ext)
		}
	}
	slices.Sort(images)
	err := galleryTemplate.Execute(os.Stdout, images)
	if err != nil {
		fatal("Okay, the template execution is being dumb:  %s\n", err)
	}
}

func pickIcon(filename string) string {
	ext := filepath.Ext(filename)
	if icon, ok := ListingIcon[ext]; ok {
		return icon
	}
	return ListingIcon["default"]
}

func sortListingEntries(slice []ListingEntry) {
	slices.SortFunc(slice, func(a, b ListingEntry) int {
		return cmp.Compare(a.Name, b.Name)
	})
}

func listing(path string) {
	listingTemplate := template.Must(template.New("listing").Parse(rawListingTemplate))
	entries := MustGetEntries(path)
	dirs := []ListingEntry{}
	files := []ListingEntry{}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		if entry.IsDir() {
			dirs = append(dirs, ListingEntry{Icon: ListingIcon["directory"], Name: entry.Name() + "/"})
		} else {
			icon := pickIcon(entry.Name())
			files = append(files, ListingEntry{Icon: icon, Name: entry.Name()})
		}
	}
	sortListingEntries(dirs)
	sortListingEntries(files)
	listing := append(dirs, files...)
	err := listingTemplate.Execute(os.Stdout, listing)
	if err != nil {
		fatal("listing template is broken: %s", err.Error())
	}
}

func main() {
	path := "./"
	flag.BoolVar(&verbosity, "verbose", false, "Make a lot of noise come out")
	flag.BoolVar(&imgMode, "images", false, "Make a tiny little image viewer")
	flag.BoolVar(&quiet, "quiet", false, "Suppress messages")
	flag.StringVar(&path, "path", "./", "Path to generate index for")
	flag.Parse()
	verbose("Verbosity active")
	if imgMode {
		message("Image gallery mode")
		gallery(path)
	} else {
		message("Directory listing mode")
		listing(path)
	}
}
