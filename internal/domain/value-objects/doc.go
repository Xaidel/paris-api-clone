// Package valueobjects contains immutable typed values used throughout the
// system. These types wrap raw strings, timestamps, identifiers, classifications,
// and pipeline metadata so the rest of the code can depend on explicit domain
// meaning instead of loosely typed primitives.
//
// Common categories in this package include:
//   - Identifiers for users, groups, transactions, uploads, and related records.
//   - Enumerated workflow states such as transaction status and classification.
//   - Structured analysis results such as step outputs, historical reuse
//     metadata, and pipeline summaries.
//   - Schema definitions and validation support for transaction file ingestion.
//
// Constructors in this package perform normalization and validation up front.
// That makes value objects a good first stop when you need to understand which
// raw inputs are considered valid and how workflow metadata is expected to look
// once it enters the domain.
package valueobjects
