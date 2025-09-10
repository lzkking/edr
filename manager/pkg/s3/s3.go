package s3

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	c1 "github.com/lzkking/edr/manager/config"
	"go.uber.org/zap"
	"io"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

var (
	S3Client *Client
)

func init() {
	var err error
	S3Client, err = NewS3Client()
	if err != nil {
		panic("连接S3失败")
	}
}

type Client struct {
	s3Bucket      string
	s3AccessKey   string
	s3SecretKey   string
	s3Region      string
	s3Endpoint    string
	s3Client      *s3.Client
	appSignSecret string
}

func (c *Client) UploadStream(ctx context.Context, r io.Reader, contentLen *int64, contentType string) (string, string, string, string, error) {
	key := fmt.Sprintf("%s", uuid.New().String())

	tmpf, err := os.CreateTemp("", "upload-*")
	if err != nil {
		return "", "", "", "", err
	}
	defer func() {
		tmpf.Close()
		os.Remove(tmpf.Name())
	}()

	hasher := sha256.New()
	tr := io.TeeReader(r, hasher)
	if _, err = io.Copy(tmpf, tr); err != nil {
		return "", "", "", "", fmt.Errorf("写入临时文件失败: %w", err)
	}
	sha256hex := hex.EncodeToString(hasher.Sum(nil))

	if _, err = tmpf.Seek(0, io.SeekStart); err != nil {
		return "", "", "", "", fmt.Errorf("临时文件seek失败: %w", err)
	}
	st, _ := tmpf.Stat()
	size := st.Size()

	in := &s3.PutObjectInput{
		Bucket: &c.s3Bucket,
		Key:    &key,
		Body:   tmpf,
	}
	if contentLen != nil {
		in.ContentLength = contentLen
	} else {
		in.ContentLength = aws.Int64(size)
	}

	if contentType != "" {
		in.ContentType = aws.String(contentType)
	}

	if _, err = c.s3Client.PutObject(ctx, in); err != nil {
		zap.S().Warnf("将文件流推送到S3失败,失败原因:%v", err)
		return "", "", "", "", err
	}

	ts := time.Now().Unix()
	signature := c.sign(key, sha256hex, ts)
	publicURL := buildPublicURL(c.s3Endpoint, c.s3Region, c.s3Bucket, key)

	return publicURL, key, sha256hex, signature, nil
}

func (c *Client) sign(key, sha256hex string, ts int64) string {
	if c.appSignSecret == "" {
		return ""
	}
	mac := hmac.New(sha256.New, []byte(c.appSignSecret))
	io.WriteString(mac, key)
	io.WriteString(mac, "|")
	io.WriteString(mac, sha256hex)
	io.WriteString(mac, "|")
	io.WriteString(mac, fmt.Sprintf("%d", ts))
	return hex.EncodeToString(mac.Sum(nil))
}

func NewS3Client() (*Client, error) {
	s3Config := c1.GetServerConfig().S3Config

	s3Client := &Client{
		s3Bucket:      s3Config.S3Bucket,
		s3AccessKey:   s3Config.S3AccessKey,
		s3SecretKey:   s3Config.S3SecretKey,
		s3Region:      s3Config.S3Region,
		s3Endpoint:    s3Config.S3Endpoint,
		s3Client:      nil,
		appSignSecret: "1342325850",
	}

	creds := credentials.NewStaticCredentialsProvider(s3Config.S3AccessKey, s3Config.S3SecretKey, "")
	var endpointResolver aws.EndpointResolverWithOptions
	if s3Config.S3Endpoint != "" {
		endpointResolver = aws.EndpointResolverWithOptionsFunc(
			func(service, region string, _ ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:           s3Config.S3Endpoint,
					SigningRegion: s3Config.S3Region,
				}, nil
			})
	}

	loadOpts := []func(*config.LoadOptions) error{
		config.WithRegion(s3Config.S3Region),
		config.WithCredentialsProvider(creds),
	}

	if endpointResolver != nil {
		loadOpts = append(loadOpts, config.WithEndpointResolverWithOptions(endpointResolver))
	}

	awsCfg, err := config.LoadDefaultConfig(context.TODO(), loadOpts...)
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if s3Config.S3Endpoint != "" {
			o.UsePathStyle = true
		}
	})

	s3Client.s3Client = client

	return s3Client, nil
}

func buildPublicURL(endpoint, region, bucket, key string) string {
	if endpoint != "" {
		base := strings.TrimRight(endpoint, "/")
		u, _ := url.Parse(base)
		u.Path = path.Join(u.Path, bucket, key)
		return u.String()
	}
	return (&url.URL{
		Scheme: "https",
		Host:   fmt.Sprintf("%s.s3.%s.amazonaws.com", bucket, region),
		Path:   path.Join("/", key),
	}).String()
}
