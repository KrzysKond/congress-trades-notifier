package congress

import (
	functions "github.com/KrzysKond/congress-trades-notifier/functions"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(string, error) {
	xml_path := functions.ZipExtractor()
	ids := functions.ProcessXML(xml_path)
	for _, id := range ids {
		pdfPath := functions.DownloadPDF(id)
		functions.UploadPDFtoS3(pdfPath, "congess-filings")
	}

}

func main() {
	lambda.Start(handler)
}
