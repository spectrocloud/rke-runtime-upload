package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io/fs"
	"os"
	"path/filepath"
)

// Creates a S3 Bucket in the region configured in the shared config
// or AWS_REGION environment variable.
//
// Usage:
//    go run s3_upload_object.go BUCKET_NAME FILENAME
func main() {
	bucket := os.Getenv("AWS_BUCKET_NAME")
	region := os.Getenv("AWS_REGION")
	//fmt.Println("AWS_ACCESS_KEY_ID : ", os.Getenv("AWS_ACCESS_KEY_ID"))
	//fmt.Println("AWS_SECRET_ACCESS_KEY : ", os.Getenv("AWS_SECRET_ACCESS_KEY"))
	//fmt.Println("AWS_BUCKET_NAME : ", os.Getenv("AWS_BUCKET_NAME"))
	//fmt.Println("AWS_REGION : ", os.Getenv("AWS_REGION"))
	//// Initialize a session in us-west-2 that the SDK will use to load
	//// credentials from the shared credentials file ~/.aws/credentials.
	k8sVersion := os.Getenv("KUBERNETES_VERSION")
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)

	svc := s3.New(sess)
	input := &s3.ListObjectsV2Input{
		Bucket:  aws.String(bucket),
		MaxKeys: aws.Int64(2),
	}

	result, err := svc.ListObjectsV2(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				fmt.Println(s3.ErrCodeNoSuchBucket, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}

	fmt.Println(result)

	//// Setup the S3 Upload Manager. Also see the SDK doc for the Upload Manager
	//// for more information on configuring part size, and concurrency.
	////
	//// http://docs.aws.amazon.com/sdk-for-go/api/service/s3/s3manager/#NewUploader
	uploader := s3manager.NewUploader(sess)

	// Upload the file's body to S3 bucket as an object with the key being the
	// same as the filename.
	dirPath := "/k8s-runtime"
	err = filepath.Walk(dirPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if info.IsDir() {
			fmt.Printf("skipping a dir without errors: %+v \n", info.Name())
			return nil
		}
		fmt.Printf("visited file or dir: %q\n", path)
		file, err := os.Open(path)
		if err != nil {
			exitErrorf("Unable to open file %q, %v", path, err)
		}
		defer file.Close()
		acl := "public-read"
		newPath := k8sVersion + path
		fmt.Printf("Uploading file %s to s3 bucket %s \n", newPath, bucket)
		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(bucket),

			// Can also use the `filepath` standard library package to modify the
			// filename as need for an S3 object key. Such as turning absolute path
			// to a relative path.
			Key: aws.String(newPath),

			// The file to be uploaded. io.ReadSeeker is preferred as the Uploader
			// will be able to optimize memory when uploading large content. io.Reader
			// is supported, but will require buffering of the reader's bytes for
			// each part.
			Body: file,
			ACL:  &acl,
		})
		if err != nil {
			// Print the error and exit.
			exitErrorf("Unable to upload %q to %q, %v", path, bucket, err)
		}

		fmt.Printf("Successfully uploaded %q to %q\n", path, bucket)
		return nil
	})
	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", dirPath, err)
		return
	}
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
