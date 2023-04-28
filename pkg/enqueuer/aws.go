package enqueuer

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/tinkeractive/transferless/pkg/job"
	"github.com/tinkeractive/transferless/pkg/transfer"
)

type AWSEnqueuer struct {
	Region        string
	JobQueue      string
	TransferQueue string
}

func NewAWSEnqueuer(region, jobQueue, transferQueue string) (AWSEnqueuer, error) {
	return AWSEnqueuer{region, jobQueue, transferQueue}, nil
}

func (e *AWSEnqueuer) EnqueueJob(job job.Job) error {
	getQueueURLInput := &sqs.GetQueueUrlInput{
		QueueName: aws.String(e.JobQueue),
	}
	sqsClient := sqs.New(session.New(), &aws.Config{
		Region: aws.String(e.Region),
	})
	getQueueURLOutput, err := sqsClient.GetQueueUrl(getQueueURLInput)
	if err != nil {
		return err
	}
	str, _ := json.Marshal(job)
	sendMessageInput := &sqs.SendMessageInput{
		MessageBody: aws.String(string(str)),
		QueueUrl:    aws.String(*getQueueURLOutput.QueueUrl),
	}
	_, err = sqsClient.SendMessage(sendMessageInput)
	if err != nil {
		return err
	}
	return nil
}

func (e *AWSEnqueuer) EnqueueTransfer(transferObj transfer.Transfer) error {
	getQueueURLInput := &sqs.GetQueueUrlInput{
		QueueName: aws.String(e.TransferQueue),
	}
	sqsClient := sqs.New(session.New(), &aws.Config{
		Region: aws.String(e.Region),
	})
	getQueueURLOutput, err := sqsClient.GetQueueUrl(getQueueURLInput)
	if err != nil {
		return err
	}
	str, _ := json.Marshal(transferObj)
	sendMessageInput := &sqs.SendMessageInput{
		MessageBody: aws.String(string(str)),
		QueueUrl:    aws.String(*getQueueURLOutput.QueueUrl),
	}
	_, err = sqsClient.SendMessage(sendMessageInput)
	if err != nil {
		return err
	}
	return nil
}
