package tests

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	functions "github.com/KrzysKond/congress-trades-notifier/functions"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ses"
)

// Mock S3 client
type mockS3Client struct{}

func (m *mockS3Client) GetObject(ctx context.Context, input *s3.GetObjectInput, opts ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	if *input.Key == "valid.txt" {
		return &s3.GetObjectOutput{
			Body: io.NopCloser(bytes.NewReader([]byte("user1@example.com\nuser2@example.com\n"))),
		}, nil
	}
	return nil, errors.New("object not found")
}

func (m *mockS3Client) PutObject(ctx context.Context, input *s3.PutObjectInput, opts ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	return &s3.PutObjectOutput{}, nil
}

// Mock SES client
type mockSESClient struct {
	SentEmails [][]byte
}

func (m *mockSESClient) SendRawEmail(ctx context.Context, input *ses.SendRawEmailInput, opts ...func(*ses.Options)) (*ses.SendRawEmailOutput, error) {
	m.SentEmails = append(m.SentEmails, input.RawMessage.Data)
	return &ses.SendRawEmailOutput{}, nil
}

func TestFetchRecipients(t *testing.T) {
	mock := &mockS3Client{}

	recipients, err := functions.FetchRecipients(mock, "bucket", "valid.txt")
	if err != nil {
		t.Fatalf("FetchRecipients failed: %v", err)
	}

	if len(recipients) != 2 {
		t.Fatalf("Expected 2 recipients, got %d", len(recipients))
	}
	if recipients[0] != "user1@example.com" || recipients[1] != "user2@example.com" {
		t.Fatalf("Recipients mismatch: %v", recipients)
	}
}

func TestSendEmails(t *testing.T) {
	mock := &mockSESClient{}
	recipients := []string{"user1@example.com", "user2@example.com"}

	err := functions.SendEmails(mock, recipients)
	if err != nil {
		t.Fatalf("SendEmails failed: %v", err)
	}

	if len(mock.SentEmails) != 2 {
		t.Fatalf("Expected 2 emails sent, got %d", len(mock.SentEmails))
	}

	for _, email := range mock.SentEmails {
		if !bytes.Contains(email, []byte("Please find the attached PDFs")) {
			t.Errorf("Email body missing expected text")
		}
	}
}
