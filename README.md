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
 - Build the go with the following command  -   **env GOOS=linux go build -o main main.go && zip deployment.zip main**

 - Upload the deployment file to lambda
 - Set the configuration

 Handler function name: main


# Configurations
You have to provide the configuration as base64 encoded format. Either add **CONFIG** in lambda variable configuration or you can pass the **config** as parameter if the lambda trigger event is like amzon API interface.

Use https://www.base64encode.org/ to encode the configurations

# Config Params

```json
{   	
  	"region" : "Region",
	"bucket" : "Source Bucket",
	"destinationbucket" : "Destination Bucket",
  	"filesources" : ["list of file sources"],
    	"tmpfilepath" : "temporary zip create path, provide if using VPC and EFS in lambda, 
			 default config will be /tmp",
    	"zipdestinationfilepath" : "Destination path for copy the ziped file within the destination buket"
  }

```
# Important
- If you are trying to zip files larger than 512 MB size, then you have to setup VPC and EFS (Elastic File System) in lambda function. Lambda only allows 512MB size to create a tmp file.
