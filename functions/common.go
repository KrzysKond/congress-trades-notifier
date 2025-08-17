package functions

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ses"
)

type S3API interface {
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

type SESAPI interface {
	SendRawEmail(ctx context.Context, params *ses.SendRawEmailInput, optFns ...func(*ses.Options)) (*ses.SendRawEmailOutput, error)
}

func Check(err error) {
	if err != nil {
		panic(err)
	}
}

func GetBucketName(s3Client *s3.Client, prefix string) string {
	resp, err := s3Client.ListBuckets(context.Background(), &s3.ListBucketsInput{})
	Check(err)

	for _, b := range resp.Buckets {
		name := *b.Name
		if strings.HasPrefix(name, prefix) && !strings.Contains(name, "logs") {
			return name
		}
	}
	panic(fmt.Sprintf("bucket not found with prefix %s", prefix))
}
