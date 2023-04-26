package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/tinkeractive/transferless/pkg/configuration"
	"github.com/tinkeractive/transferless/pkg/synchronizer"
	"github.com/tinkeractive/transferless/pkg/transfer"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/sqs"
	_ "github.com/rclone/rclone/backend/s3"
	_ "github.com/rclone/rclone/backend/sftp"
)

func main(){
	if ("" == os.Getenv("AWS_LAMBDA_FUNCTION_NAME")) {
		region := os.Getenv("AWS_REGION")
		// NOTE does not exist in deployed config
		transferQueue := os.Getenv("TRANSFERLESS_TRANSFER_QUEUE")
		sqsClient := sqs.New(session.New(), &aws.Config{
			Region: aws.String(region),
		})
		getQueueURLInput := &sqs.GetQueueUrlInput{
			QueueName: aws.String(transferQueue),
		}
		getQueueURLOutput, err := sqsClient.GetQueueUrl(getQueueURLInput)
		if err != nil {
			log.Fatal(err)
		}
		receiveMessageInput := &sqs.ReceiveMessageInput{
			QueueUrl: aws.String(*getQueueURLOutput.QueueUrl),
			MaxNumberOfMessages: aws.Int64(int64(1)),
		}
		receiveMessageOutput, err := sqsClient.ReceiveMessage(receiveMessageInput)
		if err != nil {
			log.Fatal(err)
		}
		var transferObj transfer.Transfer
		err = json.Unmarshal([]byte(*receiveMessageOutput.Messages[0].Body), &transferObj)
		if err != nil {
			log.Fatal(err)
		}
		SyncTransfer(transferObj)
	} else {
		lambda.Start(HandleRequest)
	}
}

func HandleRequest(lambdaEvent events.SQSEvent){
	for _, event := range lambdaEvent.Records {
		var transferObj transfer.Transfer
		err := json.Unmarshal([]byte(event.Body), &transferObj)
		if err != nil {
			log.Fatal(err)
		}
		SyncTransfer(transferObj)
	}
}

func SyncTransfer(transferObj transfer.Transfer) {
	log.Println("transfer:", transferObj)
	remoteConfigService := os.Getenv("TRANSFERLESS_REMOTE_CONFIG_SERVICE")
	var configProvider interface{}
	switch remoteConfigService {
	case "AWSSecretsManager":
		configProvider = configuration.AWSSecretsManager{"Type", "Transferless"}
	case "AWSSystemsManager":
		configProvider = configuration.AWSSystemsManager{"Type", "Transferless"}
	default:
		log.Println("no credential service specified")
		return
	}
	configString, err := configProvider.(configuration.Provider).GetConfig()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("loading config")
	err = configuration.LoadConfig(context.Background(), configString)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("synchronizing transfer")
	err = synchronizer.Sync(transferObj)
	if err != nil {
		log.Fatal(err)
	}
}
