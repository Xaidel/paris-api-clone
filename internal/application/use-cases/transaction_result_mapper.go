package usecases

import (
	"time"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

func newTransactionResult(transaction *entities.Transaction, step4 *entities.TransactionStep4, sector *entities.Sector, step5 *entities.TransactionStep5) outboundports.TransactionResult {
	result := outboundports.TransactionResult{
		ID:                  transaction.ID().String(),
		Product:             transaction.Product(),
		ProcessedYear:       transaction.ProcessedYear(),
		ProcessedMonth:      transaction.ProcessedMonth(),
		DMCIB:               transaction.DMCIB(),
		DMC:                 transaction.DMC(),
		PartnerBank:         transaction.PartnerBank(),
		ReferenceNumber:     transaction.ReferenceNumber(),
		TransactionValue:    transaction.TransactionValue(),
		Classification:      transaction.Classification(),
		Status:              transaction.Status(),
		PipelineResult:      newPipelineResultDetails(transaction.ClassificationValue(), transaction.PipelineResult()),
		Step4Classification: newTransactionStep4Details(step4, sector),
		Step5Classification: newTransactionStep5Details(step5),
		FailureReason:       transaction.FailureReason(),
		TransactionCount:    transaction.TransactionCount(),
		GoodsDescription:    transaction.GoodsDescription(),
		GoodsClassification: transaction.GoodsClassification(),
		ApplicantCountry:    transaction.ApplicantCountry(),
		BeneficiaryCountry:  transaction.BeneficiaryCountry(),
		SourceCountry:       transaction.SourceCountry(),
		DestinationCountry:  transaction.DestinationCountry(),
		TenorDescription:    transaction.TenorDescription(),
		ESCategory:          transaction.ESCategory(),
		PAAlignment:         transaction.PAAlignment(),
		CreatedBy:           transaction.CreatedBy(),
		CreatedAt:           transaction.CreatedAt().UTC().Format(time.RFC3339),
		UpdatedAt:           transaction.UpdatedAt().UTC().Format(time.RFC3339),
	}

	if uploadID := transaction.UploadID(); uploadID != nil {
		result.UploadID = uploadID.String()
	}

	if rowNumber := transaction.RowNumber(); rowNumber != nil {
		result.RowNumber = *rowNumber
	}

	return result
}

func newTransactionStep4Details(step4 *entities.TransactionStep4, sector *entities.Sector) *outboundports.TransactionStep4Details {
	if step4 == nil {
		return nil
	}

	identifiedSector := ""
	if sector != nil {
		identifiedSector = sector.Name()
	}

	result := "aligned"
	if step4.IsHighEmitting() {
		result = "next-step"
	}

	return &outboundports.TransactionStep4Details{
		IdentifiedSector:      identifiedSector,
		AdditionalInformation: step4.AdditionalContext().String(),
		Result:                result,
	}
}

func newTransactionStep5Details(step5 *entities.TransactionStep5) *outboundports.TransactionStep5Details {
	if step5 == nil {
		return nil
	}

	return &outboundports.TransactionStep5Details{
		ScreeningQuestion1: outboundports.TransactionStep5ScreeningQuestionDetails{
			Answer:        step5.ScreeningQuestion1Answer(),
			Justification: step5.ScreeningQuestion1Justification().String(),
		},
		ScreeningQuestion2: outboundports.TransactionStep5ScreeningQuestionDetails{
			Answer:        step5.ScreeningQuestion2Answer(),
			Justification: step5.ScreeningQuestion2Justification().String(),
		},
		ReviewerNotes: step5.ReviewerNotes().String(),
		IsFinal:       step5.IsFinal(),
		Result:        step5.Classification().String(),
	}
}

func newPipelineResultDetails(classification valueobjects.TransactionClassification, result *valueobjects.PipelineResult) *outboundports.PipelineResultDetails {
	if result == nil {
		return nil
	}

	resolvedClassification := result.FinalClassification()
	if classification.IsTerminal() {
		resolvedClassification = classification
	}

	details := &outboundports.PipelineResultDetails{
		Version:             result.Version(),
		ExitStep:            result.ExitStep(),
		FinalClassification: resolvedClassification.String(),
	}

	if legacy := result.Legacy(); legacy != nil {
		details.Step1Result = newStepResultDetails(legacy.Step1Result())
	}

	if step2Result := result.Step2Result(); step2Result != nil {
		step2Details := newStepResultDetails(*step2Result)
		details.Step2Result = &step2Details
	}

	if step3Result := result.Step3Result(); step3Result != nil {
		step3Details := newStepResultDetails(*step3Result)
		details.Step3Result = &step3Details
	}

	if react := result.React(); react != nil {
		details.ExitStep = react.ExitStep()
		reactClassification := react.OverallClassification().String()
		if classification.IsTerminal() {
			reactClassification = classification.String()
		}
		details.FinalClassification = reactClassification
		details.Reason = react.Reason()
		if batchID := react.BatchID(); batchID != nil {
			details.BatchID = *batchID
		}
		details.NotAlignedListMatch = react.NotAlignedListMatch()
		details.NotAlignedListMatchConfidence = react.NotAlignedListMatchConfidence()
		details.AlignedListMatch = react.AlignedListMatch()
		details.AlignedListMatchConfidence = react.AlignedListMatchConfidence()
	}

	return details
}

func newStepResultDetails(result valueobjects.StepResult) outboundports.StepResultDetails {
	details := outboundports.StepResultDetails{
		StepNumber:    result.StepNumber(),
		StepAlignment: result.StepAlignment().String(),
	}

	if booleanResult := result.BooleanResult(); booleanResult != nil {
		booleanResultCopy := *booleanResult
		details.BooleanResult = &booleanResultCopy
	}

	if reason := result.Reason(); reason != nil {
		details.Reason = reason.Error()
	}

	return details
}
