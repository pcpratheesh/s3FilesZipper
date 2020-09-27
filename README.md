# s3FilesZipper
Amazon lambda golang function to read s3 bucket files and create as a zip.
Move the created zip file into same or another bucket destination.

# Steps

 - Load Configuration
 - Read files from s3 related with the path provided in configuration
 - Create zip
 - Copy zip into destintion s3 bucket path

# Deployment
 - Download the deployment file from https://github.com/pratheeshpcplpta/s3FilesZipper
 - Build the go with the following command  -  env GOOS=linux go build -o main main.go && zip deployment.zip main

 - Upload the deployment file to lambda
 - Set the configuration

 Handler function name: main
