// Package services contains shared application-level coordination that is too
// cross-cutting or long-lived to live inside a single use case. These services
// still obey the application-layer rule: they orchestrate ports and domain
// objects but do not own transport details or persistence implementations.
//
// The most important groups are:
//   - Classification workers and job handlers, which dequeue queued work and
//     drive the multi-step transaction classification pipeline.
//   - Step services and pipeline services, which encapsulate reusable scoring
//     and decision flow for the classification process.
//   - Audit and event recorder services, which persist administrative activity
//     and keep event publication consistent across use cases.
//   - Embedding sync and upload progress services, which support background and
//     operational workflows that span multiple repositories.
//
// Read classification_job_handler.go and classification_pipeline_service.go
// first if you need to understand the asynchronous transaction pipeline. Read
// event_recorder_service.go and transaction_audit_service.go for the shared
// pattern used by use cases to capture administrative events.
//
// Add code here only when the logic is reused across multiple entry points or
// represents an application service with a stable responsibility. If the code is
// a single action triggered by one command or query, it usually belongs in the
// usecases package instead.
package services
