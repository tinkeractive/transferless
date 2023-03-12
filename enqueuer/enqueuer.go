package enqueuer

import (
	"github.com/tinkeractive/transferless/job"
	"github.com/tinkeractive/transferless/transfer"
)

type Enqueuer interface {
	EnqueueJob(transferJob job.Job) error
	EnqueueTransfer(transferObj transfer.Transfer) error
}
