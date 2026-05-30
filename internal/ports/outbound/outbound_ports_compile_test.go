package ports

import "testing"

func TestOutboundPortTypesCompile(t *testing.T) {
	var _ UserRepository
	var _ TransactionRepository
	var _ ClassificationJobQueue
	var _ ClassificationMetrics
}
