package storage

import (
	"errors"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/sirupsen/logrus"
)

var baseConfig = Config{
	Region:                  "eu-west-1",
	BucketVulnerableReports: "vulcan-core-vulnerable-reports-dev",
	BucketReports:           "vulcan-core-reports-dev",
	BucketLogs:              "vulcan-core-logs-dev",
	LinkBase:                "https://vulcan-results-dev.schibsted.io/v1",
}

var testCasesSaveReports = []struct {
	name             string
	skip, skipAlways bool
	config           Config
	s3Mock           s3iface.S3API
	scanID, checkID  string
	startedAt        time.Time
	report           []byte
	vulnerable       bool
	expectedLink     string
	expectedErr      bool
	output           *s3.PutObjectOutput
}{
	{
		name:   "positive-not-vulnerable",
		config: baseConfig,
		s3Mock: mockS3Client{
			putObjectOutput: &s3.PutObjectOutput{},
			err:             nil,
		},
		scanID:       "9126034c-7caf-4acd-93f3-bee1941aa140",
		checkID:      "e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0",
		startedAt:    time.Date(1984, time.April, 4, 13, 0, 0, 0, time.UTC),
		report:       []byte{},
		vulnerable:   false,
		expectedLink: "https://vulcan-results-dev.schibsted.io/v1/reports/dt=1984-04-04/scan=9126034c-7caf-4acd-93f3-bee1941aa140/e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0.json",
		expectedErr:  false,
		output:       &s3.PutObjectOutput{},
	},
	{
		name:   "positive-vulnerable",
		config: baseConfig,
		s3Mock: mockS3Client{
			putObjectOutput: &s3.PutObjectOutput{},
			err:             nil,
		},
		scanID:       "9126034c-7caf-4acd-93f3-bee1941aa140",
		checkID:      "e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0",
		startedAt:    time.Date(1984, time.April, 4, 13, 0, 0, 0, time.UTC),
		report:       []byte{},
		vulnerable:   true,
		expectedLink: "https://vulcan-results-dev.schibsted.io/v1/reports/dt=1984-04-04/scan=9126034c-7caf-4acd-93f3-bee1941aa140/e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0.json",
		expectedErr:  false,
		output:       &s3.PutObjectOutput{},
	},
	{
		name: "negative-broken-base-url",
		config: Config{
			Region:                  "eu-west-1",
			BucketVulnerableReports: "vulcan-core-vulnerable-reports-dev",
			BucketReports:           "vulcan-core-reports-dev",
			BucketLogs:              "vulcan-core-logs-dev",
			LinkBase:                "*&%$#!", // Broken base URL.
		},
		s3Mock: mockS3Client{
			putObjectOutput: &s3.PutObjectOutput{},
			err:             nil,
		},
		scanID:       "9126034c-7caf-4acd-93f3-bee1941aa140",
		checkID:      "e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0",
		startedAt:    time.Date(1984, time.April, 4, 13, 0, 0, 0, time.UTC),
		report:       []byte{},
		vulnerable:   false,
		expectedLink: "",
		expectedErr:  true,
		output:       &s3.PutObjectOutput{},
	},
	{
		name:   "negative-upload-fails",
		config: baseConfig,
		s3Mock: mockS3Client{
			putObjectOutput: &s3.PutObjectOutput{},
			err:             errors.New("error uploading file"),
		},
		scanID:       "9126034c-7caf-4acd-93f3-bee1941aa140",
		checkID:      "e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0",
		startedAt:    time.Date(1984, time.April, 4, 13, 0, 0, 0, time.UTC),
		report:       []byte{},
		vulnerable:   false,
		expectedLink: "",
		expectedErr:  true,
		output:       &s3.PutObjectOutput{},
	},
}

var testCasesSaveLogs = []struct {
	name             string
	skip, skipAlways bool
	config           Config
	s3Mock           s3iface.S3API
	scanID, checkID  string
	startedAt        time.Time
	logs             []byte
	expectedLink     string
	expectedErr      bool
	output           *s3.PutObjectOutput
}{
	{
		name:   "positive",
		config: baseConfig,
		s3Mock: mockS3Client{
			putObjectOutput: &s3.PutObjectOutput{},
			err:             nil,
		},
		scanID:       "9126034c-7caf-4acd-93f3-bee1941aa140",
		checkID:      "e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0",
		startedAt:    time.Date(1984, time.April, 4, 13, 0, 0, 0, time.UTC),
		logs:         []byte{},
		expectedLink: "https://vulcan-results-dev.schibsted.io/v1/logs/dt=1984-04-04/scan=9126034c-7caf-4acd-93f3-bee1941aa140/e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0.log",
		expectedErr:  false,
		output:       &s3.PutObjectOutput{},
	},
	{
		name: "negative-broken-base-url",
		config: Config{
			Region:                  "eu-west-1",
			BucketVulnerableReports: "vulcan-core-vulnerable-reports-dev",
			BucketReports:           "vulcan-core-reports-dev",
			BucketLogs:              "vulcan-core-logs-dev",
			LinkBase:                "*&%$#!", // Broken base URL.
		},
		s3Mock: mockS3Client{
			putObjectOutput: &s3.PutObjectOutput{},
			err:             nil,
		},
		scanID:       "9126034c-7caf-4acd-93f3-bee1941aa140",
		checkID:      "e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0",
		startedAt:    time.Date(1984, time.April, 4, 13, 0, 0, 0, time.UTC),
		logs:         []byte{},
		expectedLink: "",
		expectedErr:  true,
		output:       &s3.PutObjectOutput{},
	},
	{
		name:   "negative-upload-fails",
		config: baseConfig,
		s3Mock: mockS3Client{
			putObjectOutput: &s3.PutObjectOutput{},
			err:             errors.New("error uploading file"),
		},
		scanID:       "9126034c-7caf-4acd-93f3-bee1941aa140",
		checkID:      "e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0",
		startedAt:    time.Date(1984, time.April, 4, 13, 0, 0, 0, time.UTC),
		logs:         []byte{},
		expectedLink: "",
		expectedErr:  true,
		output:       &s3.PutObjectOutput{},
	},
}

var testCasesGetReport = []struct {
	name                  string
	config                Config
	s3Mock                s3iface.S3API
	date, scanID, checkID string
	expectedReport string
	expectedErr           bool
}{
	{
		name:   "Happy path",
		config: baseConfig,
		s3Mock: mockS3Client{
			getObjectOutput: &s3.GetObjectOutput{
				Body: ioutil.NopCloser(strings.NewReader("report")),
			},
			expectedBucket: baseConfig.BucketReports,
			expectedKey:    "dt=2019-11-16/scan=9126034c-7caf-4acd-93f3-bee1941aa140/e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0.json",
		},
		date:           "dt=2019-11-16",
		scanID:         "scan=9126034c-7caf-4acd-93f3-bee1941aa140",
		checkID:        "e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0.json",
		expectedReport: "report",
		expectedErr:    false,
	},
	{
		name:   "Should return error bad key",
		config: baseConfig,
		s3Mock: mockS3Client{
			getObjectOutput: &s3.GetObjectOutput{
				Body: ioutil.NopCloser(strings.NewReader("report")),
			},
			expectedBucket: baseConfig.BucketReports,
			expectedKey:    "dt=2019-11-16/scan=9126034c-7caf-4acd-93f3-bee1941aa140/e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0.json",
		},
		date:           "dt=2019-11-11",
		scanID:         "scan=9126034c-7caf-4acd-93f3-bee1941aa140",
		checkID:        "e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0.json",
		expectedReport: "",
		expectedErr:    true,
	},
	{
		name:   "Should return error wrong bucket",
		config: baseConfig,
		s3Mock: mockS3Client{
			getObjectOutput: &s3.GetObjectOutput{
				Body: ioutil.NopCloser(strings.NewReader("report")),
			},
			expectedBucket: "wrong bucket",
			expectedKey:    "dt=2019-11-16/scan=9126034c-7caf-4acd-93f3-bee1941aa140/e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0.json",
		},
		date:           "dt=2019-11-16",
		scanID:         "scan=9126034c-7caf-4acd-93f3-bee1941aa140",
		checkID:        "e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0.json",
		expectedReport: "",
		expectedErr:    true,
	},
}

var testCasesGetLog = []struct {
	name                  string
	config                Config
	s3Mock                s3iface.S3API
	date, scanID, checkID string
	expectedReport string
	expectedErr           bool
}{
	{
		name:   "Happy path",
		config: baseConfig,
		s3Mock: mockS3Client{
			getObjectOutput: &s3.GetObjectOutput{
				Body: ioutil.NopCloser(strings.NewReader("log")),
			},
			expectedBucket: baseConfig.BucketLogs,
			expectedKey:    "dt=2019-11-16/scan=9126034c-7caf-4acd-93f3-bee1941aa140/e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0.log",
		},
		date:           "dt=2019-11-16",
		scanID:         "scan=9126034c-7caf-4acd-93f3-bee1941aa140",
		checkID:        "e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0.log",
		expectedReport: "log",
		expectedErr:    false,
	},
	{
		name:   "Should return error bad key",
		config: baseConfig,
		s3Mock: mockS3Client{
			getObjectOutput: &s3.GetObjectOutput{
				Body: ioutil.NopCloser(strings.NewReader("log")),
			},
			expectedBucket: baseConfig.BucketLogs,
			expectedKey:    "dt=2019-11-16/scan=9126034c-7caf-4acd-93f3-bee1941aa140/e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0.log",
		},
		date:           "dt=2019-11-11",
		scanID:         "scan=9126034c-7caf-4acd-93f3-bee1941aa140",
		checkID:        "e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0.log",
		expectedReport: "",
		expectedErr:    true,
	},
	{
		name:   "Should return error wrong bucket",
		config: baseConfig,
		s3Mock: mockS3Client{
			getObjectOutput: &s3.GetObjectOutput{
				Body: ioutil.NopCloser(strings.NewReader("report")),
			},
			expectedBucket: "wrong bucket",
			expectedKey:    "dt=2019-11-16/scan=9126034c-7caf-4acd-93f3-bee1941aa140/e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0.log",
		},
		date:           "dt=2019-11-16",
		scanID:         "scan=9126034c-7caf-4acd-93f3-bee1941aa140",
		checkID:        "e0c1ac1a-1036-4e0e-b5cc-d18ae6673eb0.log",
		expectedReport: "",
		expectedErr:    true,
	},
}

type mockS3Client struct {
	s3iface.S3API
	putObjectOutput *s3.PutObjectOutput

	getObjectOutput *s3.GetObjectOutput
	expectedBucket  string
	expectedKey     string

	err error
}

func (m mockS3Client) PutObject(s *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	return m.putObjectOutput, m.err
}

func (m mockS3Client) GetObject(s *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	if *s.Bucket != m.expectedBucket || *s.Key != m.expectedKey {
		return nil, errors.New("Invalid bucket or key")
	}
	return m.getObjectOutput, m.err
}

func TestSaveReports(t *testing.T) {
	// Test all the test cases defined in testCasesSaveReports
	for _, tc := range testCasesSaveReports {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			if testing.Short() && tc.skip || tc.skipAlways {
				t.SkipNow()
			}

			l := logrus.New().WithFields(logrus.Fields{"test": tc.name})
			s := &S3Storage{Conf: tc.config, logger: l, svc: tc.s3Mock}

			link, err := s.SaveReports(tc.scanID, tc.checkID, tc.startedAt, tc.report, tc.vulnerable)
			if tc.expectedErr && err == nil {
				t.Fatalf("expected error, got none")
			} else if !tc.expectedErr && err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}

			if tc.expectedLink != "" && tc.expectedLink != link {
				t.Fatalf("expected link to %v, got: %v", tc.expectedLink, link)
			}
		})
	}
}

func TestSaveLogs(t *testing.T) {
	// Test all the test cases defined in testCasesSaveLogs
	for _, tc := range testCasesSaveLogs {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			if testing.Short() && tc.skip || tc.skipAlways {
				t.SkipNow()
			}

			l := logrus.New().WithFields(logrus.Fields{"test": tc.name})
			s := &S3Storage{Conf: tc.config, logger: l, svc: tc.s3Mock}

			link, err := s.SaveLogs(tc.scanID, tc.checkID, tc.startedAt, tc.logs)
			if tc.expectedErr && err == nil {
				t.Fatalf("expected error, got none")
			} else if !tc.expectedErr && err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}

			if tc.expectedLink != "" && tc.expectedLink != link {
				t.Fatalf("expected link to %v, got: %v", tc.expectedLink, link)
			}
		})
	}
}

func TestGetReport(t *testing.T) {
	// Test all the test cases defined in testCasesGetReport
	for _, tc := range testCasesGetReport {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			l := logrus.New().WithFields(logrus.Fields{"test": tc.name})
			s := &S3Storage{Conf: tc.config, logger: l, svc: tc.s3Mock}

			report, err := s.GetReport(tc.date, tc.scanID, tc.checkID)
			if tc.expectedErr && err == nil {
				t.Fatalf("expected error, got none")
			} else if !tc.expectedErr && err != nil {
				t.Fatalf("expected no error, got: %v", err)
			} else

			if tc.expectedReport != string(report) {
				t.Fatalf("expected report to be '%s', got: '%s'", tc.expectedReport, string(report))
			}
		})
	}
}

func TestGetLog(t *testing.T) {
	// Test all the test cases defined in testCasesGetLog
	for _, tc := range testCasesGetLog {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			l := logrus.New().WithFields(logrus.Fields{"test": tc.name})
			s := &S3Storage{Conf: tc.config, logger: l, svc: tc.s3Mock}

			log, err := s.GetLog(tc.date, tc.scanID, tc.checkID)
			if tc.expectedErr && err == nil {
				t.Fatalf("expected error, got none")
			} else if !tc.expectedErr && err != nil {
				t.Fatalf("expected no error, got: %v", err)
			} else

			if tc.expectedReport != string(log) {
				t.Fatalf("expected log to be '%s', got: '%s'", tc.expectedReport, string(log))
			}
		})
	}
}
