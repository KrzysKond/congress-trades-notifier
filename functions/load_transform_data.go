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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
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

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func ZipExtractor() string {
	const url = "https://disclosures-clerk.house.gov/public_disc/financial-pdfs/2025FD.zip"
	const zipPath = "./data/2025FD.zip"
	const XMLPath = "./data/2025FD.xml"

	// Create /data folder
	check(os.MkdirAll("./data", 0644))

	// Download ZIP
	resp, err := http.Get(url)
	check(err)
	defer resp.Body.Close()

	out, err := os.Create(zipPath)
	check(err)
	_, err = io.Copy(out, resp.Body)
	check(err)
	out.Close()

	// Open ZIP
	r, err := zip.OpenReader(zipPath)
	check(err)
	defer r.Close()

	// Extract only XML
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".xml") {
			rc, err := f.Open()
			check(err)
			outFile, err := os.Create(XMLPath)
			check(err)
			_, err = io.Copy(outFile, rc)
			check(err)
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
	check(err)
	defer file.Close()

	// Decode XML
	var fd FinancialDisclosure
	err = xml.NewDecoder(file).Decode(&fd)
	check(err)
	today := time.Now()
	yesterday := today.AddDate(0, 0, -1)

	// Filter only filings with FilingType == "P"
	var filtered []Filing
	var docIDs []string
	for _, f := range fd.Members {
		if f.FilingType == "P" {
			t, err := time.Parse("1/2/2006", f.FilingDate)
			check(err)
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
	check(err)

	// Write back to the same file
	err = os.WriteFile(xmlPath, append([]byte(xml.Header), output...), 0644)
	check(err)
	return docIDs
}

func DownloadPDF(id string) string {
	url := fmt.Sprintf("https://disclosures-clerk.house.gov/public_disc/ptr-pdfs/2025/%s.pdf", id)
	PDFPath := fmt.Sprintf("./data/%d.pdf", id)
	resp, err := http.Get(url)
	check(err)
	defer resp.Body.Close()

	out, err := os.Create(PDFPath)
	check(err)
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	check(err)

	return PDFPath
}

func UploadPDFtoS3(filePath string, bucketName string) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	check(err)

	client := s3.NewFromConfig(cfg)

	file, err := os.Open(filePath)
	check(err)
	defer file.Close()

	key := filepath.Base(filePath)

	_, err = client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		Body:   file,
		ACL:    types.ObjectCannedACLPrivate,
	})
	check(err)
}
