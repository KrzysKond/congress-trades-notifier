package functions

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	sestypes "github.com/aws/aws-sdk-go-v2/service/ses/types"
)

func FetchRecipients(client S3API, bucketName, key string) ([]string, error) {
	resp, err := client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    &key,
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var recipients []string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			recipients = append(recipients, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return recipients, nil
}

func SendEmails(client SESAPI, recipients []string) error {
	dataDir := "./data"
	files, err := os.ReadDir(dataDir)
	if err != nil {
		return err
	}

	var pdfFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".pdf") {
			pdfFiles = append(pdfFiles, filepath.Join(dataDir, f.Name()))
		}
	}

	for _, recipient := range recipients {
		var emailBody bytes.Buffer
		boundary := "MY-MULTIPART-BOUNDARY"

		emailBody.WriteString(fmt.Sprintf("To: %s\r\n", recipient))
		emailBody.WriteString("Subject: New trades by congressmen\r\n")
		emailBody.WriteString("MIME-Version: 1.0\r\n")
		emailBody.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", boundary))
		emailBody.WriteString("\r\n--" + boundary + "\r\n")
		emailBody.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
		emailBody.WriteString("Please find the attached PDFs.\r\n")

		for _, pdf := range pdfFiles {
			data, err := os.ReadFile(pdf)
			if err != nil {
				return err
			}

			emailBody.WriteString("\r\n--" + boundary + "\r\n")
			emailBody.WriteString("Content-Type: application/pdf\r\n")
			emailBody.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", filepath.Base(pdf)))
			emailBody.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
			emailBody.WriteString(encodeBase64(data))
		}

		emailBody.WriteString("\r\n--" + boundary + "--\r\n")

		_, err := client.SendRawEmail(context.Background(), &ses.SendRawEmailInput{
			RawMessage: &sestypes.RawMessage{
				Data: emailBody.Bytes(),
			},
		})
		if err != nil {
			log.Printf("Failed to send email to %s: %v", recipient, err)
		} else {
			log.Printf("Email sent to %s", recipient)
		}
	}

	return nil
}

func encodeBase64(data []byte) string {
	encoded := base64.StdEncoding.EncodeToString(data)
	var result strings.Builder
	const maxLineLength = 76
	for i := 0; i < len(encoded); i += maxLineLength {
		end := i + maxLineLength
		if end > len(encoded) {
			end = len(encoded)
		}
		result.WriteString(encoded[i:end] + "\r\n")
	}
	return result.String()
}
