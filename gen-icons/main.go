package main

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/disintegration/imaging"
)

var baseDir string = "/tmp"
var iconDir string = ""

func handler(ctx context.Context, event events.S3Event) error {
	// Get bucket and key from event
	bucket := event.Records[0].S3.Bucket.Name
	key := event.Records[0].S3.Object.Key

	// Create the source file
	tmpDir, err := ioutil.TempDir("/tmp", "")
	if err != nil {
		return fmt.Errorf("error creating tmp directory, %v", err)
	}
	baseDir = tmpDir
	iconDir = baseDir + "/icons/"

	err = os.Mkdir(iconDir, 0777)
	if err != nil {
		return fmt.Errorf("error creating icon directory, %v", err)
	}
	file, err := os.Create(baseDir + key)
	if err != nil {
		return fmt.Errorf("error creating file: %q, %v", key, err)
	}
	defer file.Close()

	// Create session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)
	if err != nil {
		return fmt.Errorf("error creating new session: %v", err)
	}

	// Obtain the source file
	downloader := s3manager.NewDownloader(sess)

	_, err = downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
	if err != nil {
		return fmt.Errorf("error downloading item: %q, %v", key, err)
	}

	// Create icons directory
	if _, err := os.Stat(iconDir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(iconDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("error creating directory %s, %v", iconDir, err)
		}
	}

	src, err := imaging.Open(file.Name())
	if err != nil {
		return fmt.Errorf("error opening image: %v", err)
	}

	// Generate icons
	err = genIcons(src)
	if err != nil {
		// Deal with it
		return fmt.Errorf("error genning icons: %v", err)
	}

	// Zip icons directory
	zipFileName := key + ".icons.zip"
	err = createZip(zipFileName)
	if err != nil {
		return fmt.Errorf("error creating zip: %v", err)
	}
	// Open generated zip file for uploading
	zipFile, err := os.Open(baseDir + zipFileName)
	if err != nil {
		return fmt.Errorf("error opening zip file: %s, %v", zipFileName, err)
	}
	defer zipFile.Close()

	// Convert to io.ReadSeeker for optimized S3 uploading
	fileInfo, err := zipFile.Stat()
	if err != nil {
		return fmt.Errorf("error stating zip file: %s, %v", zipFileName, err)
	}
	var size int64 = fileInfo.Size()

	buffer := make([]byte, size)
	zipFile.Read(buffer)
	payload := bytes.NewReader(buffer)

	// Push "Put" zip file into bucket-output
	outBucket := bucket + "-output"

	uploader := s3manager.NewUploader(sess)
	uploadResponse, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(outBucket),
		Key:    aws.String(zipFileName),
		Body:   payload,
	})
	if err != nil {
		return fmt.Errorf("unable to upload %q to %q, %v", zipFileName, bucket, err)
	}

	log.Printf("uploadResponse.Location: %s", uploadResponse.Location)

	return nil
}

func genIcons(src image.Image) error {
	sizes := []int{48, 57, 60, 72, 76, 96, 114, 120, 144, 152, 180, 192, 256, 384, 512}

	for _, size := range sizes {
		dst := imaging.Resize(src, size, size, imaging.Lanczos)
		err := imaging.Save(dst, fmt.Sprintf("%s%dx%d.jpg", iconDir, size, size))
		if err != nil {
			return err
		}
		fmt.Printf("Processed file %dx%d\n", size, size)
	}
	return nil
}

func createZip(target string) error {
	f, err := os.Create(baseDir + target)
	if err != nil {
		return err
	}
	defer f.Close()
	writer := zip.NewWriter(f)
	defer writer.Close()

	return filepath.Walk(iconDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Method = zip.Deflate

		header.Name, err = filepath.Rel(filepath.Dir(iconDir), path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			header.Name += "/"
		}

		headerWriter, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(headerWriter, f)
		return err
	})
}

func main() {
	lambda.Start(handler)
}
