# s3FilesZipper
Amazon lambda function to read s3 files and create as zip


# Steps

 - Load Configuration
 - Read files from s3 related with the path provided in configuration
 - Create zip
 - Copy zip into destintion s3 bucket path

# Deployment
 - Download the deployment file from https://github.com/pratheeshpcplpta/s3FilesZipper
 - Build the go with the following command
   env GOOS=linux go build -o s3FilesZipper s3FilesZipper.go && zip s3FilesZipper.zip s3FilesZipper

 - Upload the deployment file to lambda
 - Set the configuration

 Handler function name: main
