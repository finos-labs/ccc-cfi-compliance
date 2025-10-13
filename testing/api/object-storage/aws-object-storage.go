package objstorage

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/finos-labs/ccc-cfi-compliance/testing/api/iam"
)

// AWSS3Service implements Service for AWS S3
type AWSS3Service struct {
	client *s3.Client
	ctx    context.Context
}

// NewAWSS3Service creates a new AWS S3 service using default credentials
func NewAWSS3Service(ctx context.Context) (*AWSS3Service, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &AWSS3Service{
		client: s3.NewFromConfig(cfg),
		ctx:    ctx,
	}, nil
}

// NewAWSS3ServiceWithCredentials creates a new AWS S3 service with specific credentials from an Identity
func NewAWSS3ServiceWithCredentials(ctx context.Context, identity *iam.Identity) (*AWSS3Service, error) {
	// Extract credentials from the map
	accessKeyID := identity.Credentials["access_key_id"]
	secretAccessKey := identity.Credentials["secret_access_key"]
	sessionToken := identity.Credentials["session_token"] // Optional, empty string if not present

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKeyID,
			secretAccessKey,
			sessionToken,
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config with credentials: %w", err)
	}

	return &AWSS3Service{
		client: s3.NewFromConfig(cfg),
		ctx:    ctx,
	}, nil
}

// ListBuckets lists all S3 buckets
func (s *AWSS3Service) ListBuckets() ([]Bucket, error) {
	output, err := s.client.ListBuckets(s.ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	buckets := make([]Bucket, 0, len(output.Buckets))
	for _, b := range output.Buckets {
		buckets = append(buckets, Bucket{
			ID:     aws.ToString(b.Name),
			Name:   aws.ToString(b.Name),
			Region: "", // Region not included in list, use GetBucketRegion for specific bucket
		})
	}

	return buckets, nil
}

// CreateBucket creates a new S3 bucket in the default region
func (s *AWSS3Service) CreateBucket(bucketID string) (*Bucket, error) {
	input := &s3.CreateBucketInput{
		Bucket: aws.String(bucketID),
	}

	_, err := s.client.CreateBucket(s.ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create bucket %s: %w", bucketID, err)
	}

	// Get the actual region where the bucket was created
	region, err := s.GetBucketRegion(bucketID)
	if err != nil {
		// If we can't get region, just return without it
		region = ""
	}

	return &Bucket{
		ID:     bucketID,
		Name:   bucketID,
		Region: region,
	}, nil
}

// DeleteBucket deletes an S3 bucket
func (s *AWSS3Service) DeleteBucket(bucketID string) error {
	_, err := s.client.DeleteBucket(s.ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(bucketID),
	})
	if err != nil {
		return fmt.Errorf("failed to delete bucket %s: %w", bucketID, err)
	}

	return nil
}

// ListObjects lists all objects in a bucket
func (s *AWSS3Service) ListObjects(bucketID string) ([]Object, error) {
	output, err := s.client.ListObjectsV2(s.ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list objects in bucket %s: %w", bucketID, err)
	}

	objects := make([]Object, 0, len(output.Contents))
	for _, obj := range output.Contents {
		objects = append(objects, Object{
			ID:       aws.ToString(obj.Key),
			BucketID: bucketID,
			Name:     aws.ToString(obj.Key),
			Size:     aws.ToInt64(obj.Size),
			Data:     nil, // Don't fetch data in list operation
		})
	}

	return objects, nil
}

// CreateObject creates a new object in a bucket
func (s *AWSS3Service) CreateObject(bucketID string, objectID string, data []string) (*Object, error) {
	// Convert []string to []byte
	var content bytes.Buffer
	for _, line := range data {
		content.WriteString(line)
	}

	_, err := s.client.PutObject(s.ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketID),
		Key:    aws.String(objectID),
		Body:   bytes.NewReader(content.Bytes()),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create object %s in bucket %s: %w", objectID, bucketID, err)
	}

	return &Object{
		ID:       objectID,
		BucketID: bucketID,
		Name:     objectID,
		Size:     int64(content.Len()),
		Data:     data,
	}, nil
}

// ReadObject reads an object from a bucket
func (s *AWSS3Service) ReadObject(bucketID string, objectID string) (*Object, error) {
	output, err := s.client.GetObject(s.ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketID),
		Key:    aws.String(objectID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to read object %s from bucket %s: %w", objectID, bucketID, err)
	}
	defer output.Body.Close()

	// Read the content
	content, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object content: %w", err)
	}

	return &Object{
		ID:       objectID,
		BucketID: bucketID,
		Name:     objectID,
		Size:     aws.ToInt64(output.ContentLength),
		Data:     []string{string(content)},
	}, nil
}

// DeleteObject deletes an object from a bucket
func (s *AWSS3Service) DeleteObject(bucketID string, objectID string) error {
	_, err := s.client.DeleteObject(s.ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucketID),
		Key:    aws.String(objectID),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object %s from bucket %s: %w", objectID, bucketID, err)
	}

	return nil
}

// GetBucketRegion gets the region where a bucket is located
func (s *AWSS3Service) GetBucketRegion(bucketID string) (string, error) {
	output, err := s.client.GetBucketLocation(s.ctx, &s3.GetBucketLocationInput{
		Bucket: aws.String(bucketID),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get bucket location for %s: %w", bucketID, err)
	}

	// AWS returns empty string for us-east-1
	region := string(output.LocationConstraint)
	if region == "" {
		region = "us-east-1"
	}

	return region, nil
}
