package tests

import (
	"encoding/xml"
	"os"
	"testing"

	functions "github.com/KrzysKond/congress-trades-notifier/functions"
)

func TestExtractZipFile(t *testing.T) {
	xmlPath := functions.ZipExtractor()

	info, err := os.Stat(xmlPath)
	if err != nil {
		t.Fatalf("XML doesn't exist: %v", err)
	}

	if info.Size() == 0 {
		t.Fatal("XML file empty")
	}
	os.Remove(xmlPath)
}

func TestProcessXML(t *testing.T) {
	xmlPath := functions.ZipExtractor()
	functions.ProcessXML(xmlPath)
	info, err := os.Stat(xmlPath)
	if err != nil {
		t.Fatalf("XML doesn't exist: %v", err)
	}

	if info.Size() == 0 {
		t.Fatal("XML file empty")
	}

	data, err := os.ReadFile(xmlPath)
	if err != nil {
		t.Fatalf("Failed to read XML: %v", err)
	}

	var fd functions.FinancialDisclosure
	err = xml.Unmarshal(data, &fd)
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	for _, f := range fd.Members {
		if f.FilingType != "P" {
			t.Fatalf("Found a filing with wrong type: %s", f.FilingType)
		}
	}

	os.Remove(xmlPath)
}

func TestDownloadPDF(t *testing.T) {
	const sampleID = "20030630" // can be different but has to be a real one, type P
	pdfFile := functions.DownloadPDF(sampleID)

	info, err := os.Stat(pdfFile)
	if err != nil {
		t.Fatalf("PDF doesn't exist: %v", err)
	}

	if info.Size() == 0 {
		t.Fatal("PDF file empty")
	}
	os.Remove(pdfFile)
}
