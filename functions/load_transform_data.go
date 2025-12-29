package functions

import (
	"archive/zip"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type Filing struct {
	Prefix     string `xml:"Prefix"`
	Last       string `xml:"Last"`
	First      string `xml:"First"`
	Suffix     string `xml:"Suffix"`
	FilingType string `xml:"FilingType"`
	StateDst   string `xml:"StateDst"`
	Year       string `xml:"Year"`
	FilingDate string `xml:"FilingDate"`
	DocID      string `xml:"DocID"`
}

type FinancialDisclosure struct {
	Members []Filing `xml:"Member"`
}

func ZipExtractor() string {
	year := time.Now().Year()
	url := fmt.Sprintf(
		"https://disclosures-clerk.house.gov/public_disc/financial-pdfs/%dFD.zip",
		year,
	)
	zipPath := fmt.Sprintf("/tmp/lambda/%dFD.zip", year)
	XMLPath := fmt.Sprintf("/tmp/lambda/%dFD.xml", year)

	// Create /lambda folder
	Check(os.MkdirAll("/tmp/lambda", 0755))

	// Download ZIP
	resp, err := http.Get(url)
	Check(err)
	defer resp.Body.Close()

	out, err := os.Create(zipPath)
	Check(err)
	_, err = io.Copy(out, resp.Body)
	Check(err)
	out.Close()

	// Open ZIP
	r, err := zip.OpenReader(zipPath)
	Check(err)
	defer r.Close()

	// Extract only XML
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".xml") {
			rc, err := f.Open()
			Check(err)
			outFile, err := os.Create(XMLPath)
			Check(err)
			_, err = io.Copy(outFile, rc)
			Check(err)
			rc.Close()
			outFile.Close()
			break
		}
	}

	// Cleanup ZIP
	os.Remove(zipPath)

	return XMLPath
}

func ProcessXML(xmlPath string) []string {
	file, err := os.Open(xmlPath)
	Check(err)
	defer file.Close()

	// Decode XML
	var fd FinancialDisclosure
	err = xml.NewDecoder(file).Decode(&fd)
	Check(err)
	today := time.Now()
	yesterday := today.AddDate(0, 0, -1)

	// Filter only filings with FilingType == "P"
	var filtered []Filing
	var docIDs []string
	for _, f := range fd.Members {
		if f.FilingType == "P" {
			t, err := time.Parse("1/2/2006", f.FilingDate)
			Check(err)
			if t.Year() == today.Year() && t.YearDay() == today.YearDay() ||
				t.Year() == yesterday.Year() && t.YearDay() == yesterday.YearDay() {
				f.FilingDate = t.Format("2/1/2006") // DD/MM/YYYY
				filtered = append(filtered, f)
				docIDs = append(docIDs, f.DocID)
			}
		}
	}

	// Replace Members with filtered list
	fd.Members = filtered

	// Marshal filtered XML
	output, err := xml.MarshalIndent(fd, "", "  ")
	Check(err)

	// Write back to the same file
	err = os.WriteFile(xmlPath, append([]byte(xml.Header), output...), 0644)
	Check(err)
	return docIDs
}

func DownloadPDF(id string) string {
	url := fmt.Sprintf("https://disclosures-clerk.house.gov/public_disc/ptr-pdfs/2025/%s.pdf", id)
	PDFPath := fmt.Sprintf("/tmp/lambda/%d.pdf", id)
	resp, err := http.Get(url)
	Check(err)
	defer resp.Body.Close()

	out, err := os.Create(PDFPath)
	Check(err)
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	Check(err)

	return PDFPath
}

func UploadPDFtoS3(client S3API, filePath string, bucketName string) error {
	key := filepath.Base(filePath)
	fmt.Printf("[DEBUG] UploadPDFtoS3 called with bucket=%s, key=%s\n", bucketName, key)

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("[ERROR] Failed to open file %s: %v\n", filePath, err)
		return err
	}
	defer file.Close()

	_, err = client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: &bucketName,
		Key:    &key,
		Body:   file,
		ACL:    types.ObjectCannedACLPrivate,
	})
	if err != nil {
		fmt.Printf("[ERROR] Failed to PutObject: %v\n", err)
		return err
	}

	fmt.Printf("[DEBUG] Successfully uploaded %s to bucket %s\n", key, bucketName)
	return nil
}
