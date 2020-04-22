package storage

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"net/url"
	"path"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/sirupsen/logrus"
)

// Config represents the configuration options for S3Storage objects
type Config struct {
	BucketVulnerableReports string
	BucketReports           string
	BucketLogs              string
	Region                  string
	LinkBase                string
	Endpoint                string
	PathStyle               bool
}

// Storage is an interface of a type that can save a result.
type Storage interface {
	SaveReports(scanID, checkID string, startedAt time.Time, report []byte, vulnerable bool) (link string, err error)
	SaveLogs(scanID, checkID string, startedAt time.Time, logs []byte) (link string, err error)

	GetReport(date, scanID, checkID string) ([]byte, error)
	GetLog(date, scanID, checkID string) ([]byte, error)
}

// S3Storage implements the Storage interface storing the results in S3.
type S3Storage struct {
	Conf   Config
	logger *logrus.Entry
	svc    s3iface.S3API
}

// NewS3Storage creates a S3Storage for a specified bucket.
func NewS3Storage(c Config, l *logrus.Entry, s s3iface.S3API) *S3Storage {
	return &S3Storage{Conf: c, logger: l, svc: s}
}

// SaveReports stores the result in an S3 file.
func (s *S3Storage) SaveReports(scanID, checkID string, startedAt time.Time, report []byte, vulnerable bool) (link string, err error) {
	//see http://docs.aws.amazon.com/athena/latest/ug/partitions.html
	dt := startedAt.Format("dt=2006-01-02")
	scan := "scan=" + scanID

	key := fmt.Sprintf("%s/%s/%s.json", dt, scan, checkID)
	if vulnerable {
		compress := true
		err = s.uploadToBucket(s.Conf.BucketVulnerableReports, key+".gz", report, compress, aws.String("gzip"))
		if err != nil {
			return "", err
		}
	}

	link, err = urlConcat(s.Conf.LinkBase, "reports", dt, scan, checkID+".json")
	if err != nil {
		return "", err
	}

	err = s.uploadToBucket(s.Conf.BucketReports, key, report, false, aws.String("text/json"))
	if err != nil {
		return "", err
	}

	return link, err
}

// SaveLogs stores the result in an S3 file.
func (s *S3Storage) SaveLogs(scanID, checkID string, startedAt time.Time, logs []byte) (link string, err error) {
	//see http://docs.aws.amazon.com/athena/latest/ug/partitions.html
	dt := startedAt.Format("dt=2006-01-02")
	scan := "scan=" + scanID

	key := fmt.Sprintf("%s/%s/%s.log", dt, scan, checkID)

	link, err = urlConcat(s.Conf.LinkBase, "logs", dt, scan, checkID+".log")
	if err != nil {
		return "", err
	}

	err = s.uploadToBucket(s.Conf.BucketLogs, key, logs, false, nil)
	if err != nil {
		return "", err
	}

	return
}

// GetReport downloads from S3 and returns the report that corresponds
// to the input params.
func (s *S3Storage) GetReport(date, scanID, checkID string) ([]byte, error) {
	key := fmt.Sprintf("%s/%s/%s", date, scanID, checkID)

	return s.downloadFromBucket(s.Conf.BucketReports, key)
}

// GetLog downloads from S3 and returns the report that corresponds
// to the input params.
func (s *S3Storage) GetLog(date, scanID, checkID string) ([]byte, error) {
	key := fmt.Sprintf("%s/%s/%s", date, scanID, checkID)

	return s.downloadFromBucket(s.Conf.BucketLogs, key)
}

func (s *S3Storage) uploadToBucket(bucket, key string, content []byte, compress bool, contentType *string) (err error) {
	if compress {
		var buf bytes.Buffer
		zw := gzip.NewWriter(&buf)
		_, err = zw.Write(content)
		if err != nil {
			return err
		}

		if err = zw.Close(); err != nil {
			return err
		}
		content = buf.Bytes()
	}

	s.logger.WithFields(logrus.Fields{
		"content": string(content),
		"key":     key,
		"bucket":  bucket,
	}).Debug("uploading content to S3 bucket")

	params := &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(content),
		ContentType: contentType,
	}
	_, err = s.svc.PutObject(params)

	return
}

func (s *S3Storage) downloadFromBucket(bucket, key string) ([]byte, error) {
	s.logger.WithFields(logrus.Fields{
		"key":    key,
		"bucket": bucket,
	}).Debug("downloading content from S3 bucket")

	params := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	obj, err := s.svc.GetObject(params)
	if err != nil {
		return nil, err
	}
	defer obj.Body.Close()

	return ioutil.ReadAll(obj.Body)
}

func urlConcat(baseURL string, toConcat ...string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	toJoin := append([]string{u.Path}, toConcat...)
	u.Path = path.Join(toJoin...)

	return u.String(), nil
}
