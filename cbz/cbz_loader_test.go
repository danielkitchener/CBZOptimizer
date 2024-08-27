package cbz

import (
	"strings"
	"testing"
)

func TestLoadChapter(t *testing.T) {
	// Define the path to the .cbz file
	chapterFilePath := "../testdata/Chapter 1.cbz"

	// Load the chapter
	chapter, err := LoadChapter(chapterFilePath)
	if err != nil {
		t.Fatalf("Failed to load chapter: %v", err)
	}

	// Check the number of pages
	expectedPages := 16
	actualPages := len(chapter.Pages)
	if actualPages != expectedPages {
		t.Errorf("Expected %d pages, but got %d", expectedPages, actualPages)
	}

	// Check if ComicInfoXml contains the expected series name
	expectedSeries := "<Series>Boundless Necromancer</Series>"
	if !strings.Contains(chapter.ComicInfoXml, expectedSeries) {
		t.Errorf("ComicInfoXml does not contain the expected series: %s", expectedSeries)
	}
}
