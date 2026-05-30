package usecases

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gyud-adb/paris-api/internal/domain"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

func buildTransactionFilter(query inboundports.ListTransactionsQuery) (outboundports.TransactionFilter, error) {
	filter := outboundports.TransactionFilter{
		SortBy:    outboundports.TransactionSortByCreatedAt,
		SortOrder: outboundports.TransactionSortOrderDescending,
	}

	if query.UploadID != "" {
		uploadID, err := valueobjects.UploadIDFromString(query.UploadID)
		if err != nil {
			return outboundports.TransactionFilter{}, fmt.Errorf("parsing upload id: %w", err)
		}
		filter.UploadID = &uploadID
	}

	createdAtFrom, err := parseTransactionDateFilter("created_at_from", query.CreatedAtFrom)
	if err != nil {
		return outboundports.TransactionFilter{}, err
	}
	filter.CreatedAtFrom = createdAtFrom

	createdAtTo, err := parseTransactionDateFilter("created_at_to", query.CreatedAtTo)
	if err != nil {
		return outboundports.TransactionFilter{}, err
	}
	filter.CreatedAtTo = createdAtTo

	if filter.CreatedAtFrom != nil && filter.CreatedAtTo != nil && filter.CreatedAtFrom.After(*filter.CreatedAtTo) {
		return outboundports.TransactionFilter{}, domain.NewValidationError([]domain.FieldValidationError{
			domain.NewFieldValidationError("created_at_from", "invalid_range", "created_at_from must be less than or equal to created_at_to"),
		})
	}

	assignTransactionStringFilter(&filter.ApplicantCountry, query.ApplicantCountry)
	assignTransactionStringFilter(&filter.BeneficiaryCountry, query.BeneficiaryCountry)
	assignTransactionStringFilter(&filter.SourceCountry, query.SourceCountry)
	assignTransactionStringFilter(&filter.DestinationCountry, query.DestinationCountry)

	transactionCountMin, err := parseTransactionCountFilter("transaction_count_min", query.TransactionCountMin)
	if err != nil {
		return outboundports.TransactionFilter{}, err
	}
	filter.TransactionCountMin = transactionCountMin

	transactionCountMax, err := parseTransactionCountFilter("transaction_count_max", query.TransactionCountMax)
	if err != nil {
		return outboundports.TransactionFilter{}, err
	}
	filter.TransactionCountMax = transactionCountMax

	if filter.TransactionCountMin != nil && filter.TransactionCountMax != nil && *filter.TransactionCountMin > *filter.TransactionCountMax {
		return outboundports.TransactionFilter{}, domain.NewValidationError([]domain.FieldValidationError{
			domain.NewFieldValidationError("transaction_count_min", "invalid_range", "transaction_count_min must be less than or equal to transaction_count_max"),
		})
	}

	if query.Classification != "" {
		classification, err := valueobjects.TransactionClassificationFromString(query.Classification)
		if err != nil {
			return outboundports.TransactionFilter{}, fmt.Errorf("parsing classification: %w", err)
		}
		value := classification.String()
		filter.Classification = &value
	}

	if query.Status != "" {
		status, err := valueobjects.TransactionStatusFromString(query.Status)
		if err != nil {
			return outboundports.TransactionFilter{}, fmt.Errorf("parsing status: %w", err)
		}
		value := status.String()
		filter.Status = &value
	}

	sortBy, err := normalizeTransactionSortBy(query.SortBy)
	if err != nil {
		return outboundports.TransactionFilter{}, err
	}
	filter.SortBy = sortBy

	sortOrder, err := normalizeTransactionSortOrder(query.SortOrder)
	if err != nil {
		return outboundports.TransactionFilter{}, err
	}
	filter.SortOrder = sortOrder

	return filter, nil
}

func parseTransactionDateFilter(field, raw string) (*time.Time, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	parsed, err := time.Parse(time.DateOnly, strings.TrimSpace(raw))
	if err != nil {
		return nil, domain.NewValidationError([]domain.FieldValidationError{
			domain.NewFieldValidationError(field, "invalid_value", field+" must use YYYY-MM-DD format"),
		})
	}

	return &parsed, nil
}

func parseTransactionCountFilter(field, raw string) (*int, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return nil, domain.NewValidationError([]domain.FieldValidationError{
			domain.NewFieldValidationError(field, "invalid_value", field+" must be an integer"),
		})
	}

	if value < 0 {
		return nil, domain.NewValidationError([]domain.FieldValidationError{
			domain.NewFieldValidationError(field, "invalid_value", field+" must be greater than or equal to 0"),
		})
	}

	return &value, nil
}

func assignTransactionStringFilter(target **string, raw string) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return
	}

	*target = &trimmed
}

func normalizeTransactionSortBy(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return outboundports.TransactionSortByCreatedAt, nil
	}

	switch strings.ToLower(strings.TrimSpace(raw)) {
	case outboundports.TransactionSortByCreatedAt:
		return outboundports.TransactionSortByCreatedAt, nil
	case outboundports.TransactionSortByApplicantCountry:
		return outboundports.TransactionSortByApplicantCountry, nil
	case outboundports.TransactionSortByBeneficiaryCountry:
		return outboundports.TransactionSortByBeneficiaryCountry, nil
	case outboundports.TransactionSortBySourceCountry:
		return outboundports.TransactionSortBySourceCountry, nil
	case outboundports.TransactionSortByDestinationCountry:
		return outboundports.TransactionSortByDestinationCountry, nil
	case outboundports.TransactionSortByTransactionCount:
		return outboundports.TransactionSortByTransactionCount, nil
	case outboundports.TransactionSortByClassification:
		return outboundports.TransactionSortByClassification, nil
	case outboundports.TransactionSortByStatus:
		return outboundports.TransactionSortByStatus, nil
	default:
		return "", domain.NewValidationError([]domain.FieldValidationError{
			domain.NewFieldValidationError("sort_by", "invalid_value", "sort_by is invalid"),
		})
	}
}

func normalizeTransactionSortOrder(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return outboundports.TransactionSortOrderDescending, nil
	}

	switch strings.ToLower(strings.TrimSpace(raw)) {
	case outboundports.TransactionSortOrderAscending:
		return outboundports.TransactionSortOrderAscending, nil
	case outboundports.TransactionSortOrderDescending:
		return outboundports.TransactionSortOrderDescending, nil
	default:
		return "", domain.NewValidationError([]domain.FieldValidationError{
			domain.NewFieldValidationError("sort_order", "invalid_value", "sort_order must be asc or desc"),
		})
	}
}

func normalizeNavigationClassification(raw string) (*string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}

	if strings.EqualFold(trimmed, "needs review") {
		value := valueobjects.NextStepTransactionClassification().String()
		return &value, nil
	}

	classification, err := valueobjects.TransactionClassificationFromString(trimmed)
	if err != nil {
		return nil, err
	}

	value := classification.String()
	return &value, nil
}

func normalizeNavigationStep(raw string) (*int, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || strings.EqualFold(trimmed, "all") {
		return nil, nil
	}

	value, err := strconv.Atoi(trimmed)
	if err != nil || value < 1 || value > 5 {
		return nil, domain.NewValidationError([]domain.FieldValidationError{
			domain.NewFieldValidationError("step", "invalid_value", "step must be one of: 1, 2, 3, 4, 5"),
		})
	}

	return &value, nil
}
