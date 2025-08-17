package main

import (
	"context"

	functions "github.com/KrzysKond/congress-trades-notifier/functions"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ses"
)

func handler(ctx context.Context) {
	cfg, err := config.LoadDefaultConfig(ctx)
	functions.Check(err)

	s3Client := s3.NewFromConfig(cfg)
	sesClient := ses.NewFromConfig(cfg)

	xmlPath := functions.ZipExtractor()
	docIDs := functions.ProcessXML(xmlPath)
	bucketName := functions.GetBucketName(s3Client, "cg-fillings")
	for _, id := range docIDs {
		pdfPath := functions.DownloadPDF(id)
		functions.UploadPDFtoS3(s3Client, pdfPath, bucketName)
	}

	recipients, err := functions.FetchRecipients(s3Client, bucketName, "receipients.txt")
	functions.Check(err)
	functions.SendEmails(sesClient, recipients)
}

func main() {
	lambda.Start(handler)
}
