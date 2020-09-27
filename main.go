/**
 *
 * @Author : PRATHEESH PC
 * @Mail   : stackframedeveloper@gmail.com
 *
 * @Description
 * This is a aws lambda function for
 * 	 - read the files from s3
 * 	 - make a zip
 * 	 - move the zipped files into a destination bucket
 *
 * @Links
 *
 * Learn about lambda
 * https://www.tutorialspoint.com/aws_lambda/aws_lambda_overview.htm
 */
package main

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"io/ioutil"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type Config struct {
	ZipDestinationFilePath string `json:"zipdestinationfilepath"`
	ZipFileName            string `json:"zipfilename"`
	TmpFilePath            string `json:"tmpfilepath"`

	Region            string `json:"region"`
	Bucket            string `json:"bucket"`
	DestinationBucket string `json:"destinationbucket"`

	FileSources []interface{} `json:"filesources"`
}

type ZipperResponse struct {
	FilePath          string `json:"filepath"`
	FileName          string `json:"filename"`
	DestinationBucket string `json:"destinationbucket"`
	Status            bool   `json:"ok"`
	Message           string `json:"Message"`
}

var config Config

// load configuration from env
// configuration in format of base64 encode
func parseConfig(payload map[string]interface{}) (err error) {
	data := make([]byte, 0)
	config = Config{}
	configVar := os.Getenv("CONFIG")
	// check the input have params
	if payload != nil {
		if ok := payload["config"]; ok != nil {
			configVar = payload["config"].(string)
		}
	}

	data, err = base64.StdEncoding.DecodeString(configVar)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return fmt.Errorf("no configuration available")
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		return err
	}

	t := time.Now()
	if config.ZipDestinationFilePath == "" {
		config.ZipDestinationFilePath = "lambda_ziped/"
	}

	config.ZipFileName = "backup_file_on_" + t.Format("2006-01-02__15:04:05") + strconv.Itoa(int(time.Now().UnixNano())) + ".zip"

	if config.TmpFilePath == "" {
		config.TmpFilePath = "/tmp/"
	}
	return nil
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func Handler(event interface{}) (ZipperResponse, error) {

	// load events params
	e := event.(map[string]interface{})
	payLoad := e["payload"]
	var payLoadMap map[string]interface{}
	if payLoad != nil {
		payLoadMap = payLoad.(map[string]interface{})
	}

	//load configuration
	err := parseConfig(payLoadMap)
	fmt.Println("Config : ", config.Bucket)
	if err != nil {
		return ZipperResponse{
			Status:  false,
			Message: fmt.Sprintf("error in parsing the config %v", err),
		}, fmt.Errorf("error in parsing the config %v", err)
	}

	//create a session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(config.Region)},
	)

	if err != nil {
		return ZipperResponse{
			Status:  false,
			Message: fmt.Sprintf("Unable to create session %v", err),
		}, fmt.Errorf("Unable to create session %v", err)
	}

	if err != nil {
		return ZipperResponse{
			Status:  false,
			Message: fmt.Sprintf("Unable to process dynamo connection %v", err),
		}, fmt.Errorf("Unable to process dynamo connection %v", err)
	}

	svc := s3.New(sess)

	//generate zip file from the given path
	err = generateZip(svc)

	if err != nil {
		return ZipperResponse{
			Status:  false,
			Message: fmt.Sprintf("Unable to create or zip files %v", err),
		}, fmt.Errorf("Unable to create or zip files %v", err)
	}

	filename := config.TmpFilePath
	file, err := os.Open(filename)
	defer file.Close()

	if err != nil {
		return ZipperResponse{
			Status:  false,
			Message: fmt.Sprintf("Unable to open file %v", err),
		}, fmt.Errorf("Unable to open file %v", err)
	}

	// destination bucket
	var destinationBucket string
	if exists := config.DestinationBucket; exists != "" {
		destinationBucket = config.DestinationBucket
	} else {
		destinationBucket = config.Bucket
	}

	// upload the created zip file into destination bucket
	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(destinationBucket),
		Key:    aws.String(config.ZipDestinationFilePath + config.ZipFileName),
		Body:   file,
	})
	return ZipperResponse{
		FilePath:          config.ZipDestinationFilePath,
		FileName:          config.ZipFileName,
		DestinationBucket: destinationBucket,
		Status:            true,
		Message:           "success",
	}, err
}

// Generate zip file
func generateZip(svc *s3.S3) error {

	// prepare zip file
	filename := config.TmpFilePath
	file, err := ioutil.TempFile(filename, "prefix-")
	if err != nil {
		return fmt.Errorf("Unable to zip file in folder %v", err)
	}
	config.TmpFilePath = file.Name()

	// Create a new zip archive
	zipWriter := zip.NewWriter(file)
	//closing the file after all the write completes
	defer zipWriter.Close()

	// fetch the files from list of sources
	for _, part := range config.FileSources {

		//list objects within bucket
		var continuationToken string
		for {
			_object := &s3.ListObjectsV2Input{
				Bucket:  aws.String(config.Bucket),
				Prefix:  aws.String(part.(string)),
				MaxKeys: aws.Int64(10000),
			}
			if continuationToken != "" {
				_object.ContinuationToken = aws.String(continuationToken)
			}

			resp, err := svc.ListObjectsV2(_object)

			if err != nil {
				exitErrorf("Unable to list items in bucket %q, %v", config.Bucket, err)
				break
			}

			for _, item := range resp.Contents {
				fmt.Printf("Processing '%s'\n", *item.Key)

				if strings.Contains(*item.Key, config.ZipDestinationFilePath) {
					continue
				}

				ziped_file_name := filepath.Base(*item.Key)
				zipFile, err := zipWriter.Create(ziped_file_name)

				out, err := svc.GetObject(&s3.GetObjectInput{
					Bucket: aws.String(config.Bucket),
					Key:    aws.String(*item.Key),
				})

				buf := new(bytes.Buffer)
				buf.ReadFrom(out.Body)

				_, err = zipFile.Write(buf.Bytes())
				fmt.Println(err)

			}

			if resp.NextContinuationToken != nil {
				continuationToken = *resp.NextContinuationToken
			} else {
				break
			}
		}

	}

	return nil
}

func main() {
	lambda.Start(Handler)
}
