package ports

// ClassificationMetrics records classification pipeline observability events.
type ClassificationMetrics interface {
	IncPipelineTotal(status string)
	IncStepExitTotal(step string, alignment string)
	ObservePipelineDurationSeconds(seconds float64)
	IncConfidenceStubTotal()
	IncReactHistoricalExactMatchTotal()
	IncReactLLMRouteTotal()
	ObserveReactLLMBatchSize(size int)
	ObserveReactLLMLatencySeconds(seconds float64)
	IncReactInvalidLLMOutputTotal()
	IncReactClassificationTotal(classification string)
	IncReactStatusTotal(status string)
}
