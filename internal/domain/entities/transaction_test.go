package entities

import (
	"errors"
	"testing"
	"time"

	"github.com/gyud-adb/paris-api/internal/domain"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// This test verifies the standalone constructor normalizes a valid transaction
// into the default processing workflow state with matching timestamps.
func TestNewTransaction(t *testing.T) {
	t.Parallel()

	transactionID, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	transaction, err := NewTransaction(
		transactionID,
		"CG",
		2026,
		4,
		"IB",
		"DMC",
		"Partner Bank",
		"REF-1",
		"698436.80",
		1,
		"Goods",
		"Classification",
		"Philippines",
		"Japan",
		"Thailand",
		"Philippines",
		"N",
		"",
		"PA Aligned",
		time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("NewTransaction() error = %v", err)
	}

	if transaction.ID().String() != transactionID.String() {
		t.Fatalf("transaction.ID() = %q, want %q", transaction.ID().String(), transactionID.String())
	}

	if transaction.UploadID() != nil {
		t.Fatal("transaction.UploadID() != nil, want nil")
	}

	if transaction.Product() != "CG" {
		t.Fatalf("transaction.Product() = %q, want %q", transaction.Product(), "CG")
	}

	if transaction.ProcessedYear() != 2026 {
		t.Fatalf("transaction.ProcessedYear() = %d, want %d", transaction.ProcessedYear(), 2026)
	}

	if transaction.ReferenceNumber() != "REF-1" {
		t.Fatalf("transaction.ReferenceNumber() = %q, want %q", transaction.ReferenceNumber(), "REF-1")
	}

	if transaction.TransactionValue() != "698436.80" {
		t.Fatalf("transaction.TransactionValue() = %q, want %q", transaction.TransactionValue(), "698436.80")
	}

	if transaction.Classification() != "unclassified" {
		t.Fatalf("transaction.Classification() = %q, want %q", transaction.Classification(), "unclassified")
	}

	if transaction.Status() != "processing" {
		t.Fatalf("transaction.Status() = %q, want %q", transaction.Status(), "processing")
	}

	if transaction.PAAlignment() != "PA Aligned" {
		t.Fatalf("transaction.PAAlignment() = %q, want %q", transaction.PAAlignment(), "PA Aligned")
	}

	if !transaction.UpdatedAt().Equal(transaction.CreatedAt()) {
		t.Fatalf("transaction.UpdatedAt() = %v, want %v", transaction.UpdatedAt(), transaction.CreatedAt())
	}
}

// This test verifies uploaded rows retain upload-specific provenance while
// still starting in the same default unclassified processing state.
func TestNewUploadedTransaction(t *testing.T) {
	t.Parallel()

	transactionID, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	uploadID, err := valueobjects.UploadIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
	if err != nil {
		t.Fatalf("UploadIDFromString() error = %v", err)
	}

	transaction, err := NewUploadedTransaction(
		transactionID,
		uploadID,
		2,
		"CG",
		2026,
		4,
		"IB",
		"DMC",
		"Partner Bank",
		"REF-1",
		"698436.80",
		1,
		"Goods",
		"Classification",
		"Philippines",
		"Japan",
		"Thailand",
		"Philippines",
		"N",
		"",
		"PA Aligned",
		time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("NewUploadedTransaction() error = %v", err)
	}

	if transaction.UploadID() == nil || transaction.UploadID().String() != uploadID.String() {
		t.Fatalf("transaction.UploadID() = %v, want %q", transaction.UploadID(), uploadID.String())
	}

	if transaction.RowNumber() == nil || *transaction.RowNumber() != 2 {
		t.Fatalf("transaction.RowNumber() = %v, want %d", transaction.RowNumber(), 2)
	}

	if transaction.ProcessedYear() != 2026 {
		t.Fatalf("transaction.ProcessedYear() = %d, want %d", transaction.ProcessedYear(), 2026)
	}

	if transaction.Classification() != "unclassified" {
		t.Fatalf("transaction.Classification() = %q, want %q", transaction.Classification(), "unclassified")
	}

	if transaction.Status() != "processing" {
		t.Fatalf("transaction.Status() = %q, want %q", transaction.Status(), "processing")
	}
}

// This test guards the mutable-update path so all persisted fields, review
// state, and timestamps move together when an existing transaction is edited.
func TestTransactionUpdate(t *testing.T) {
	t.Parallel()

	transactionID, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	transaction, err := NewTransaction(
		transactionID,
		"CG",
		2026,
		4,
		"IB",
		"DMC",
		"Partner Bank",
		"REF-1",
		"698436.80",
		1,
		"Goods",
		"Classification",
		"Philippines",
		"Japan",
		"Thailand",
		"Philippines",
		"N",
		"",
		"PA Aligned",
		time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("NewTransaction() error = %v", err)
	}

	updatedAt := time.Date(2026, time.April, 3, 12, 0, 0, 0, time.UTC)
	if err := transaction.Update("Updated Product", 2027, 5, "Updated IB", "Updated DMC", "Updated Bank", "REF-2", "712000.25", "aligned", "ai-reviewed", 2, "Updated Goods", "Updated Classification", "Singapore", "Japan", "Malaysia", "Singapore", "Y", "A", "Misaligned", updatedAt); err != nil {
		t.Fatalf("transaction.Update() error = %v", err)
	}

	if transaction.Classification() != "aligned" {
		t.Fatalf("transaction.Classification() = %q, want %q", transaction.Classification(), "aligned")
	}

	if transaction.Status() != "ai-reviewed" {
		t.Fatalf("transaction.Status() = %q, want %q", transaction.Status(), "ai-reviewed")
	}

	if transaction.Product() != "Updated Product" {
		t.Fatalf("transaction.Product() = %q, want %q", transaction.Product(), "Updated Product")
	}

	if transaction.ProcessedYear() != 2027 {
		t.Fatalf("transaction.ProcessedYear() = %d, want %d", transaction.ProcessedYear(), 2027)
	}

	if transaction.TransactionValue() != "712000.25" {
		t.Fatalf("transaction.TransactionValue() = %q, want %q", transaction.TransactionValue(), "712000.25")
	}

	if transaction.TransactionCount() != 2 {
		t.Fatalf("transaction.TransactionCount() = %d, want %d", transaction.TransactionCount(), 2)
	}

	if transaction.GoodsDescription() != "Updated Goods" {
		t.Fatalf("transaction.GoodsDescription() = %q, want %q", transaction.GoodsDescription(), "Updated Goods")
	}

	if !transaction.UpdatedAt().Equal(updatedAt) {
		t.Fatalf("transaction.UpdatedAt() = %v, want %v", transaction.UpdatedAt(), updatedAt)
	}
}

// TestTransactionMarkProcessing verifies the transaction mark processing behavior and the expected outcome asserted below.
func TestTransactionMarkProcessing(t *testing.T) {
	t.Parallel()

	transactionID, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	transaction, err := NewTransaction(transactionID, "CG", 2026, 4, "IB", "DMC", "Partner Bank", "REF-1", "698436.80", 1, "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "N", "", "PA Aligned", time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("NewTransaction() error = %v", err)
	}

	updatedAt := time.Date(2026, time.April, 3, 12, 0, 0, 0, time.UTC)
	if err := transaction.MarkProcessing(updatedAt); err != nil {
		t.Fatalf("transaction.MarkProcessing() error = %v", err)
	}

	if transaction.Status() != "processing" {
		t.Fatalf("transaction.Status() = %q, want %q", transaction.Status(), "processing")
	}

	if !transaction.UpdatedAt().Equal(updatedAt) {
		t.Fatalf("transaction.UpdatedAt() = %v, want %v", transaction.UpdatedAt(), updatedAt)
	}
}

// TestTransactionMarkClassified verifies the transaction mark classified behavior and the expected outcome asserted below.
func TestTransactionMarkClassified(t *testing.T) {
	t.Parallel()

	transactionID, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	transaction, err := NewTransaction(transactionID, "CG", 2026, 4, "IB", "DMC", "Partner Bank", "REF-1", "698436.80", 1, "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "N", "", "PA Aligned", time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("NewTransaction() error = %v", err)
	}

	updatedAt := time.Date(2026, time.April, 3, 12, 0, 0, 0, time.UTC)
	step1Result := valueobjects.NewStepResult(
		1,
		valueobjects.NewMatchDecision(valueobjects.AlignedAlignment(), 0.9, "u2", nil),
		valueobjects.NewMatchDecision(valueobjects.UnalignedAlignment(), 0.4, "u2", nil),
	)
	pipelineResult := valueobjects.NewPipelineResult(
		transaction.ID().String(),
		step1Result,
		nil,
		nil,
		1,
		valueobjects.NotAlignedTransactionClassification(),
	)

	if err := transaction.MarkClassified(valueobjects.NotAlignedTransactionClassification(), pipelineResult, updatedAt); err != nil {
		t.Fatalf("transaction.MarkClassified() error = %v", err)
	}

	if transaction.Classification() != "not-aligned" {
		t.Fatalf("transaction.Classification() = %q, want %q", transaction.Classification(), "not-aligned")
	}

	if transaction.Status() != "ai-reviewed" {
		t.Fatalf("transaction.Status() = %q, want %q", transaction.Status(), "ai-reviewed")
	}

	if transaction.PipelineResult() == nil {
		t.Fatal("transaction.PipelineResult() = nil, want pipeline result")
	}

	if transaction.FailureReason() != "" {
		t.Fatalf("transaction.FailureReason() = %q, want empty", transaction.FailureReason())
	}

	if !transaction.UpdatedAt().Equal(updatedAt) {
		t.Fatalf("transaction.UpdatedAt() = %v, want %v", transaction.UpdatedAt(), updatedAt)
	}

	stored := transaction.PipelineResult()
	if stored == nil || stored.ExitStep() != 1 {
		t.Fatalf("transaction.PipelineResult().ExitStep() = %v, want %d", stored, 1)
	}

	if !stored.FinalClassification().Equal(valueobjects.NotAlignedTransactionClassification()) {
		t.Fatalf("transaction.PipelineResult().FinalClassification() = %q, want %q", stored.FinalClassification().String(), valueobjects.NotAlignedTransactionClassification().String())
	}
}

// TestTransactionMarkClassifiedAllowsNextStep verifies the transaction mark classified allows next step behavior and the expected outcome asserted below.
func TestTransactionMarkClassifiedAllowsNextStep(t *testing.T) {
	t.Parallel()

	transactionID, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	transaction, err := NewTransaction(transactionID, "CG", 2026, 4, "IB", "DMC", "Partner Bank", "REF-1", "698436.80", 1, "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "Y", "", "PA Aligned", time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("NewTransaction() error = %v", err)
	}

	updatedAt := time.Date(2026, time.April, 3, 12, 0, 0, 0, time.UTC)
	step1 := valueobjects.NewStepResult(
		1,
		valueobjects.NewMatchDecision(valueobjects.UnalignedAlignment(), 0.1, "u1", nil),
		valueobjects.NewMatchDecision(valueobjects.UnalignedAlignment(), 0.1, "u1", nil),
	)
	step2 := valueobjects.NewStepResult(
		2,
		valueobjects.NewMatchDecision(valueobjects.UnalignedAlignment(), 0.1, "sector", nil),
		valueobjects.NewMatchDecision(valueobjects.UnalignedAlignment(), 0.1, "sector", nil),
	)
	step3 := valueobjects.NewBooleanStepResult(3, false)
	pipelineResult := valueobjects.NewPipelineResult(
		transaction.ID().String(),
		step1,
		&step2,
		&step3,
		3,
		valueobjects.NextStepTransactionClassification(),
	)

	if err := transaction.MarkClassified(valueobjects.NextStepTransactionClassification(), pipelineResult, updatedAt); err != nil {
		t.Fatalf("transaction.MarkClassified() error = %v", err)
	}

	if transaction.Classification() != valueobjects.NextStepTransactionClassification().String() {
		t.Fatalf("transaction.Classification() = %q, want %q", transaction.Classification(), valueobjects.NextStepTransactionClassification().String())
	}

	if transaction.Status() != valueobjects.AIReviewedTransactionStatus().String() {
		t.Fatalf("transaction.Status() = %q, want %q", transaction.Status(), valueobjects.AIReviewedTransactionStatus().String())
	}
}

// TestTransactionMarkClassifiedFromPreviousTransaction verifies the transaction mark classified from previous transaction behavior and the expected outcome asserted below.
func TestTransactionMarkClassifiedFromPreviousTransaction(t *testing.T) {
	t.Parallel()

	transactionID, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	transaction, err := NewTransaction(transactionID, "CG", 2026, 4, "IB", "DMC", "Partner Bank", "REF-1", "698436.80", 1, "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "N", "", "PA Aligned", time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("NewTransaction() error = %v", err)
	}

	reactResult := valueobjects.NewReactPipelineResult(valueobjects.NewReactReviewResult(
		valueobjects.PipelineResultSourceExactHistoricalMatch,
		"react",
		"v1",
		"v1",
		"gpt-4o-mini",
		transaction.ID().String(),
		nil,
		nil,
		nil,
		nil,
		true,
		10,
		false,
		0,
		1,
		valueobjects.NotAlignedTransactionClassification(),
		"reused from previous transaction",
	))

	updatedAt := time.Date(2026, time.April, 3, 12, 0, 0, 0, time.UTC)
	if err := transaction.MarkClassifiedFromPreviousTransaction(valueobjects.NotAlignedTransactionClassification(), reactResult, updatedAt); err != nil {
		t.Fatalf("transaction.MarkClassifiedFromPreviousTransaction() error = %v", err)
	}

	if transaction.Status() != "from-previous-transactions" {
		t.Fatalf("transaction.Status() = %q, want %q", transaction.Status(), "from-previous-transactions")
	}

	if transaction.PipelineResult() == nil || transaction.PipelineResult().React() == nil {
		t.Fatal("transaction.PipelineResult().React() = nil, want react result")
	}
}

// TestTransactionMarkClassifiedFromPreviousTransactionRejectsNextStep verifies the transaction mark classified from previous transaction rejects next step behavior and the expected outcome asserted below.
func TestTransactionMarkClassifiedFromPreviousTransactionRejectsNextStep(t *testing.T) {
	t.Parallel()

	transactionID, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	transaction, err := NewTransaction(transactionID, "CG", 2026, 4, "IB", "DMC", "Partner Bank", "REF-1", "698436.80", 1, "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "Y", "", "PA Aligned", time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("NewTransaction() error = %v", err)
	}

	reactResult := valueobjects.NewReactPipelineResult(valueobjects.NewReactReviewResult(
		valueobjects.PipelineResultSourceExactHistoricalMatch,
		"react",
		"v1",
		"v1",
		"gpt-4o-mini",
		transaction.ID().String(),
		nil,
		nil,
		nil,
		nil,
		false,
		0,
		false,
		0,
		3,
		valueobjects.NextStepTransactionClassification(),
		"review pending",
	))

	err = transaction.MarkClassifiedFromPreviousTransaction(valueobjects.NextStepTransactionClassification(), reactResult, time.Date(2026, time.April, 3, 12, 0, 0, 0, time.UTC))
	if !errors.Is(err, domain.ErrInvalidTransactionClassification) {
		t.Fatalf("transaction.MarkClassifiedFromPreviousTransaction() error = %v, want %v", err, domain.ErrInvalidTransactionClassification)
	}
}

// TestTransactionMarkProfessionallyReviewed verifies the transaction mark professionally reviewed behavior and the expected outcome asserted below.
func TestTransactionMarkProfessionallyReviewed(t *testing.T) {
	t.Parallel()

	transactionID, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	transaction, err := NewTransaction(transactionID, "CG", 2026, 4, "IB", "DMC", "Partner Bank", "REF-1", "698436.80", 1, "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "Y", "", "PA Aligned", time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("NewTransaction() error = %v", err)
	}

	updatedAt := time.Date(2026, time.April, 3, 12, 0, 0, 0, time.UTC)
	if err := transaction.MarkProfessionallyReviewed(valueobjects.AlignedTransactionClassification(), updatedAt); err != nil {
		t.Fatalf("transaction.MarkProfessionallyReviewed() error = %v", err)
	}

	if transaction.Classification() != valueobjects.AlignedTransactionClassification().String() {
		t.Fatalf("transaction.Classification() = %q, want %q", transaction.Classification(), valueobjects.AlignedTransactionClassification().String())
	}

	if transaction.Status() != valueobjects.ProfessionallyReviewedTransactionStatus().String() {
		t.Fatalf("transaction.Status() = %q, want %q", transaction.Status(), valueobjects.ProfessionallyReviewedTransactionStatus().String())
	}

	if !transaction.UpdatedAt().Equal(updatedAt) {
		t.Fatalf("transaction.UpdatedAt() = %v, want %v", transaction.UpdatedAt(), updatedAt)
	}
}

// TestTransactionMarkProfessionallyReviewedAllowsNextStep verifies the transaction mark professionally reviewed allows next step behavior and the expected outcome asserted below.
func TestTransactionMarkProfessionallyReviewedAllowsNextStep(t *testing.T) {
	t.Parallel()

	transactionID, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	transaction, err := NewTransaction(transactionID, "CG", 2026, 4, "IB", "DMC", "Partner Bank", "REF-1", "698436.80", 1, "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "Y", "", "PA Aligned", time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("NewTransaction() error = %v", err)
	}

	updatedAt := time.Date(2026, time.April, 3, 12, 0, 0, 0, time.UTC)
	if err := transaction.MarkProfessionallyReviewed(valueobjects.NextStepTransactionClassification(), updatedAt); err != nil {
		t.Fatalf("transaction.MarkProfessionallyReviewed() error = %v", err)
	}

	if transaction.Classification() != valueobjects.NextStepTransactionClassification().String() {
		t.Fatalf("transaction.Classification() = %q, want %q", transaction.Classification(), valueobjects.NextStepTransactionClassification().String())
	}

	if transaction.Status() != valueobjects.ProfessionallyReviewedTransactionStatus().String() {
		t.Fatalf("transaction.Status() = %q, want %q", transaction.Status(), valueobjects.ProfessionallyReviewedTransactionStatus().String())
	}
}

// TestTransactionHasStep3NextStepResult verifies the transaction has step 3 next step result behavior and the expected outcome asserted below.
func TestTransactionHasStep3NextStepResult(t *testing.T) {
	t.Parallel()

	transactionID, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	tests := []struct {
		name             string
		pipelineResult   *valueobjects.PipelineResult
		wantEligibleStep bool
	}{
		{
			name: "legacy step 3 next step",
			pipelineResult: pipelineResultPointer(valueobjects.NewPipelineResult(
				transactionID.String(),
				valueobjects.NewStepResult(
					1,
					valueobjects.NewMatchDecision(valueobjects.UnalignedAlignment(), 0.1, "u1", nil),
					valueobjects.NewMatchDecision(valueobjects.UnalignedAlignment(), 0.1, "u1", nil),
				),
				stepResultPointer(valueobjects.NewStepResult(
					2,
					valueobjects.NewMatchDecision(valueobjects.UnalignedAlignment(), 0.1, "sector", nil),
					valueobjects.NewMatchDecision(valueobjects.UnalignedAlignment(), 0.1, "sector", nil),
				)),
				stepResultPointer(valueobjects.NewBooleanStepResult(3, false)),
				3,
				valueobjects.NextStepTransactionClassification(),
			)),
			wantEligibleStep: true,
		},
		{
			name: "react step 3 next step",
			pipelineResult: pipelineResultPointer(valueobjects.NewReactPipelineResult(valueobjects.NewReactReviewResult(
				valueobjects.PipelineResultSourceLegacyStep3Fallback,
				"react",
				"v1",
				"v1",
				"gpt-4o-mini",
				transactionID.String(),
				nil,
				nil,
				nil,
				nil,
				false,
				1,
				false,
				1,
				3,
				valueobjects.NextStepTransactionClassification(),
				"review required",
			))),
			wantEligibleStep: true,
		},
		{
			name: "step 3 aligned is not eligible",
			pipelineResult: pipelineResultPointer(valueobjects.NewPipelineResult(
				transactionID.String(),
				valueobjects.NewStepResult(
					1,
					valueobjects.NewMatchDecision(valueobjects.UnalignedAlignment(), 0.1, "u1", nil),
					valueobjects.NewMatchDecision(valueobjects.UnalignedAlignment(), 0.1, "u1", nil),
				),
				stepResultPointer(valueobjects.NewStepResult(
					2,
					valueobjects.NewMatchDecision(valueobjects.UnalignedAlignment(), 0.1, "sector", nil),
					valueobjects.NewMatchDecision(valueobjects.UnalignedAlignment(), 0.1, "sector", nil),
				)),
				stepResultPointer(valueobjects.NewBooleanStepResult(3, true)),
				3,
				valueobjects.AlignedTransactionClassification(),
			)),
			wantEligibleStep: false,
		},
		{name: "missing pipeline result is not eligible", wantEligibleStep: false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			transaction := ReconstituteTransaction(
				transactionID,
				nil,
				nil,
				"CG",
				2026,
				4,
				"IB",
				"DMC",
				"Partner Bank",
				"REF-1",
				"698436.80",
				valueobjects.UnclassifiedTransactionClassification(),
				valueobjects.ProcessingTransactionStatus(),
				tc.pipelineResult,
				"",
				1,
				"Goods",
				"Classification",
				"Philippines",
				"Japan",
				"Thailand",
				"Philippines",
				"Y",
				"",
				"PA Aligned",
				valueobjects.UserID{},
				time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC),
				time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC),
			)

			got := transaction.HasStep3NextStepResult()
			if got != tc.wantEligibleStep {
				t.Fatalf("transaction.HasStep3NextStepResult() = %v, want %v", got, tc.wantEligibleStep)
			}
		})
	}
}

func stepResultPointer(result valueobjects.StepResult) *valueobjects.StepResult {
	resultCopy := result
	return &resultCopy
}

func pipelineResultPointer(result valueobjects.PipelineResult) *valueobjects.PipelineResult {
	resultCopy := result
	return &resultCopy
}

// TestTransactionMarkFailed verifies the transaction mark failed behavior and the expected outcome asserted below.
func TestTransactionMarkFailed(t *testing.T) {
	t.Parallel()

	transactionID, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	transaction, err := NewTransaction(transactionID, "CG", 2026, 4, "IB", "DMC", "Partner Bank", "REF-1", "698436.80", 1, "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "N", "", "PA Aligned", time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("NewTransaction() error = %v", err)
	}

	step1Result := valueobjects.NewStepResult(
		1,
		valueobjects.NewMatchDecision(valueobjects.AlignedAlignment(), 0.9, "u2", nil),
		valueobjects.NewMatchDecision(valueobjects.UnalignedAlignment(), 0.4, "u2", nil),
	)
	pipelineResult := valueobjects.NewPipelineResult(
		transaction.ID().String(),
		step1Result,
		nil,
		nil,
		1,
		valueobjects.NotAlignedTransactionClassification(),
	)
	if err := transaction.MarkClassified(valueobjects.NotAlignedTransactionClassification(), pipelineResult, time.Date(2026, time.April, 3, 12, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("transaction.MarkClassified() error = %v", err)
	}

	updatedAt := time.Date(2026, time.April, 4, 12, 0, 0, 0, time.UTC)
	if err := transaction.MarkFailed("pipeline exploded", updatedAt); err != nil {
		t.Fatalf("transaction.MarkFailed() error = %v", err)
	}

	if transaction.Status() != "failed" {
		t.Fatalf("transaction.Status() = %q, want %q", transaction.Status(), "failed")
	}

	if transaction.FailureReason() != "pipeline exploded" {
		t.Fatalf("transaction.FailureReason() = %q, want %q", transaction.FailureReason(), "pipeline exploded")
	}

	if transaction.PipelineResult() != nil {
		t.Fatal("transaction.PipelineResult() != nil, want nil")
	}

	if !transaction.UpdatedAt().Equal(updatedAt) {
		t.Fatalf("transaction.UpdatedAt() = %v, want %v", transaction.UpdatedAt(), updatedAt)
	}

	if err := transaction.MarkClassified(valueobjects.AlignedTransactionClassification(), pipelineResult, time.Date(2026, time.April, 5, 12, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("transaction.MarkClassified() after failure error = %v", err)
	}

	if transaction.FailureReason() != "" {
		t.Fatalf("transaction.FailureReason() after classify = %q, want empty", transaction.FailureReason())
	}
}

// TestNewTransactionValidationErrors verifies the new transaction validation errors behavior and the expected outcome asserted below.
func TestNewTransactionValidationErrors(t *testing.T) {
	t.Parallel()

	transactionID, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	_, err = NewTransaction(transactionID, "", 0, 13, "", "", "", "", "bad-value", -1, "", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "N", "", "", time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC))
	if err == nil {
		t.Fatal("NewTransaction() error = nil, want validation error")
	}

	var validationErr *domain.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %v", err)
	}

	if len(validationErr.Fields()) != 8 {
		t.Fatalf("len(validationErr.Fields()) = %d, want %d", len(validationErr.Fields()), 8)
	}
}
