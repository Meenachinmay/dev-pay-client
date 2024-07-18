package handlers

import (
	"context"
	payments "dev-pay-client/grpcproto"
	. "github.com/tigerbeetle/tigerbeetle-go/pkg/types"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

func (config *Config) CreateTransferHandler(successCount *int64, failCount *int64, wg *sync.WaitGroup) {
	transferClient := payments.NewCreateTransferServiceClient(config.Client)

	// jobs for create txns
	createTxnsJobs := make(chan []Transfer, REQUEST_COUNT/BATCH_SIZE)

	// start worker goroutines for create transfer client
	for i := 0; i < CONCURRENCY; i++ {
		wg.Add(1)
		go workerCreateTransfer(transferClient, createTxnsJobs, successCount, failCount, wg)
	}

	//send jobs for creating transfers
	for i := 0; i < REQUEST_COUNT; i += BATCH_SIZE {
		end := i + BATCH_SIZE
		if end > REQUEST_COUNT {
			end = REQUEST_COUNT
		}
		batch := generateTransfers(end - i)
		createTxnsJobs <- batch
	}
	close(createTxnsJobs)
}

func generateTransfers(count int) []Transfer {
	transfers := make([]Transfer, count)
	for i := 0; i < count; i++ {
		fromID := uint64(i%2 + 1)   // Alternates between 1 and 2
		toID := uint64((i % 2) + 3) // Alternates between 3 and 4

		transfers[i] = Transfer{
			ID:              ToUint128(uint64(i + 1)),
			DebitAccountID:  ToUint128(fromID),
			CreditAccountID: ToUint128(toID),
			Amount:          ToUint128(5), // 5 USD
			Ledger:          1,
			Code:            1,
		}
	}
	return transfers
}

func convertBatchToProto(batch []Transfer) []*payments.Transfer {
	protoTransfers := make([]*payments.Transfer, len(batch))
	for i, t := range batch {
		protoTransfers[i] = convertTBTransferToProtoTransfer(t)
	}
	return protoTransfers
}

func workerCreateTransfer(client payments.CreateTransferServiceClient, jobs <-chan []Transfer, successCount, failCount *int64, wg *sync.WaitGroup) {
	defer wg.Done()

	for batch := range jobs {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		req := &payments.CreateTransfersRequest{Transfers: convertBatchToProto(batch)}
		_, err := client.CreateTransfers(ctx, req)
		cancel()

		if err != nil {
			log.Printf("Failed to create batch of transfers: %v", err)
			atomic.AddInt64(failCount, int64(len(batch)))
		} else {
			atomic.AddInt64(successCount, int64(len(batch)))
		}
	}
}

func convertTBTransferToProtoTransfer(t Transfer) *payments.Transfer {
	return &payments.Transfer{
		Id: uint64(t.ID.Bytes()[0]) | uint64(t.ID.Bytes()[1])<<8 | uint64(t.ID.Bytes()[2])<<16 | uint64(t.ID.Bytes()[3])<<24 |
			uint64(t.ID.Bytes()[4])<<32 | uint64(t.ID.Bytes()[5])<<40 | uint64(t.ID.Bytes()[6])<<48 | uint64(t.ID.Bytes()[7])<<56,
		DebitAccountId: uint64(t.DebitAccountID.Bytes()[0]) | uint64(t.DebitAccountID.Bytes()[1])<<8 | uint64(t.DebitAccountID.Bytes()[2])<<16 | uint64(t.DebitAccountID.Bytes()[3])<<24 |
			uint64(t.DebitAccountID.Bytes()[4])<<32 | uint64(t.DebitAccountID.Bytes()[5])<<40 | uint64(t.DebitAccountID.Bytes()[6])<<48 | uint64(t.DebitAccountID.Bytes()[7])<<56,
		CreditAccountId: uint64(t.CreditAccountID.Bytes()[0]) | uint64(t.CreditAccountID.Bytes()[1])<<8 | uint64(t.CreditAccountID.Bytes()[2])<<16 | uint64(t.CreditAccountID.Bytes()[3])<<24 |
			uint64(t.CreditAccountID.Bytes()[4])<<32 | uint64(t.CreditAccountID.Bytes()[5])<<40 | uint64(t.CreditAccountID.Bytes()[6])<<48 | uint64(t.CreditAccountID.Bytes()[7])<<56,
		Amount: uint64(t.Amount.Bytes()[0]) | uint64(t.Amount.Bytes()[1])<<8 | uint64(t.Amount.Bytes()[2])<<16 | uint64(t.Amount.Bytes()[3])<<24 |
			uint64(t.Amount.Bytes()[4])<<32 | uint64(t.Amount.Bytes()[5])<<40 | uint64(t.Amount.Bytes()[6])<<48 | uint64(t.Amount.Bytes()[7])<<56,
		Ledger: t.Ledger,
		Code:   uint32(t.Code),
	}
}
