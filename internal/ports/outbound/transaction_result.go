package ports

// PipelineResultDetails exposes the persisted transaction review-result envelope.
type PipelineResultDetails struct {
	Version                       string
	BatchID                       string
	NotAlignedListMatch           bool
	NotAlignedListMatchConfidence int
	AlignedListMatch              bool
	AlignedListMatchConfidence    int
	Reason                        string
	Step1Result                   StepResultDetails
	Step2Result                   *StepResultDetails
	Step3Result                   *StepResultDetails
	ExitStep                      int
	FinalClassification           string
}

// StepResultDetails exposes one pipeline step outcome.
type StepResultDetails struct {
	StepNumber    int
	StepAlignment string
	BooleanResult *bool
	Reason        string
}

// TransactionStep4Details exposes the persisted step 4 review details.
type TransactionStep4Details struct {
	IdentifiedSector      string
	AdditionalInformation string
	Result                string
}

// TransactionStep5ScreeningQuestionDetails exposes one persisted step 5 screening answer.
type TransactionStep5ScreeningQuestionDetails struct {
	Answer        bool
	Justification string
}

// TransactionStep5Details exposes the persisted step 5 screening details.
type TransactionStep5Details struct {
	ScreeningQuestion1 TransactionStep5ScreeningQuestionDetails
	ScreeningQuestion2 TransactionStep5ScreeningQuestionDetails
	ReviewerNotes      *string
	IsFinal            bool
	Result             string
}

// TransactionResult exposes one persisted transaction record.
type TransactionResult struct {
	ID                  string
	UploadID            string
	RowNumber           int
	Product             string
	ProcessedYear       int
	ProcessedMonth      int
	DMCIB               string
	DMC                 string
	PartnerBank         string
	ReferenceNumber     string
	TransactionValue    string
	Classification      string
	Status              string
	PipelineResult      *PipelineResultDetails
	Step4Classification *TransactionStep4Details
	Step5Classification *TransactionStep5Details
	FailureReason       string
	TransactionCount    int
	GoodsDescription    string
	GoodsClassification string
	ApplicantCountry    string
	BeneficiaryCountry  string
	SourceCountry       string
	DestinationCountry  string
	TenorDescription    string
	ESCategory          string
	PAAlignment         string
	CreatedBy           string
	CreatedAt           string
	UpdatedAt           string
}
