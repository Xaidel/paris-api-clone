package observability

import outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"

// NoopClassificationMetrics is a placeholder metrics sink until a concrete
// metrics backend is wired.
type NoopClassificationMetrics struct{}

// NewClassificationMetrics builds the classification metrics collector.
//
// TODO: replace this no-op stub with the project's metrics backend once one is
// introduced.
func NewClassificationMetrics() outboundports.ClassificationMetrics {
	return &NoopClassificationMetrics{}
}

// IncPipelineTotal records one pipeline completion by status.
func (m *NoopClassificationMetrics) IncPipelineTotal(status string) {
	_ = m
	_ = status
}

// IncStepExitTotal records one pipeline exit by step and alignment.
func (m *NoopClassificationMetrics) IncStepExitTotal(step string, alignment string) {
	_ = m
	_ = step
	_ = alignment
}

// ObservePipelineDurationSeconds records pipeline runtime.
func (m *NoopClassificationMetrics) ObservePipelineDurationSeconds(seconds float64) {
	_ = m
	_ = seconds
}

// IncConfidenceStubTotal records one stub confidence output.
func (m *NoopClassificationMetrics) IncConfidenceStubTotal() {
	_ = m
}

// IncReactHistoricalExactMatchTotal records one exact historical ReAct reuse hit.
func (m *NoopClassificationMetrics) IncReactHistoricalExactMatchTotal() {
	_ = m
}

// IncReactLLMRouteTotal records one unresolved route to the ReAct LLM.
func (m *NoopClassificationMetrics) IncReactLLMRouteTotal() {
	_ = m
}

// ObserveReactLLMBatchSize records one ReAct LLM batch size.
func (m *NoopClassificationMetrics) ObserveReactLLMBatchSize(size int) {
	_ = m
	_ = size
}

// ObserveReactLLMLatencySeconds records one ReAct LLM latency measurement.
func (m *NoopClassificationMetrics) ObserveReactLLMLatencySeconds(seconds float64) {
	_ = m
	_ = seconds
}

// IncReactInvalidLLMOutputTotal records one invalid ReAct output.
func (m *NoopClassificationMetrics) IncReactInvalidLLMOutputTotal() {
	_ = m
}

// IncReactClassificationTotal records one final ReAct classification.
func (m *NoopClassificationMetrics) IncReactClassificationTotal(classification string) {
	_ = m
	_ = classification
}

// IncReactStatusTotal records one final ReAct status.
func (m *NoopClassificationMetrics) IncReactStatusTotal(status string) {
	_ = m
	_ = status
}
