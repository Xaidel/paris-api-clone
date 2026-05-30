// Package services contains domain algorithms that are part of the business
// model but do not belong to a single entity. These services consume value
// objects and entities, apply business-oriented calculations or validations, and
// return domain-friendly results without performing I/O themselves.
//
// The package currently centers on transaction-file validation, normalization,
// and similarity helpers used by the classification workflow to compare
// transaction text against curated reference data.
//
// Read transaction_file_validator.go when working on input schemas and the
// similarity helpers when changing scoring or comparison rules. If a feature
// needs repositories, queues, loggers, or HTTP awareness, it belongs in
// application or adapters rather than here.
package services
