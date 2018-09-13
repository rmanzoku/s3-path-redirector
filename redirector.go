package redirector

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Redirector is redirector struct
type Redirector struct {
	RedirectToFormat string
	Bucket           string
	StatePrefix      string
	LinkPrefix       string
	Region           string
	S3               *s3.S3
}

// NewRedirector return *Redirector
func NewRedirector() (*Redirector, error) {
	ret := new(Redirector)
	ret.RedirectToFormat = "%s"
	ret.StatePrefix = "state/"
	ret.LinkPrefix = ""

	return ret, nil
}

// Prepare sets up for S3
func (r *Redirector) Prepare() error {

	sess, err := session.NewSession()
	if err != nil {
		return err
	}

	config := aws.NewConfig()
	config = config.WithRegion(r.Region)

	r.S3 = s3.New(sess, config)

	return nil
}

// CreateLink creates link conditionally
func (r Redirector) CreateLink(key string) (string, error) {
	linkPath, err := r.getState(key)
	if err != nil {
		return "", err
	}

	// Link is already exists
	if linkPath != "" {
		return linkPath, nil
	}

	for {
		linkPath = r.createLinkPath()
		b, err := r.linkPathExists(linkPath)

		if err != nil {
			return "", err
		}

		if b {
			continue
		}

		err = r.setState(key, linkPath)
		if err != nil {
			return "", err
		}
		err = r.createLinkFile(key, linkPath)
		if err != nil {
			return "", err
		}

		break
	}

	return linkPath, nil
}

// GetRedirectToURI returns redirect uri
func (r Redirector) GetRedirectToURI(redirectKey string) (string, error) {
	s := fmt.Sprintf(r.RedirectToFormat, redirectKey)
	return s, nil
}

func (r Redirector) createLinkFile(key, linkPath string) error {
	input := &s3.PutObjectInput{
		Bucket: aws.String(r.Bucket),
		ACL:    aws.String("public-read"),
	}
	input.Key = aws.String(linkPath)
	redirect, _ := r.GetRedirectToURI(key)
	input.WebsiteRedirectLocation = aws.String(redirect)

	return r.putObject(input)
}

func (r Redirector) linkPathExists(linkPath string) (bool, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(r.Bucket),
		Key:    aws.String(linkPath),
	}

	_, err := r.S3.HeadObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "NotFound":
				return false, nil
			default:
				return true, err
			}
		} else {
			return true, err
		}
	}
	return true, err
}

func (r Redirector) getState(key string) (string, error) {
	input := &s3.GetObjectInput{}
	input.Bucket = aws.String(r.Bucket)
	input.Key = aws.String(r.StatePrefix + key)

	result, err := r.S3.GetObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				return "", nil
			default:
				return "", err
			}
		}
	}

	b, err := ioutil.ReadAll(result.Body)
	return string(b), err
}

func (r Redirector) setState(key, linkPath string) error {
	input := &s3.PutObjectInput{
		Bucket: aws.String(r.Bucket),
		ACL:    aws.String("private"),
	}

	input.Body = bytes.NewReader([]byte(linkPath))
	input.Key = aws.String(r.StatePrefix + key)

	return r.putObject(input)
}

func (r Redirector) putObject(input *s3.PutObjectInput) error {
	_, err := r.S3.PutObject(input)
	return err
}

func (r Redirector) createLinkPath() string {
	return r.LinkPrefix + randString(6)
}

func randString(n int) string {
	rs1Letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = rs1Letters[rand.Intn(len(rs1Letters))]
	}
	return string(b)
}
