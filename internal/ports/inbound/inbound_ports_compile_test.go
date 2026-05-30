package ports

import "testing"

func TestInboundPortTypesCompile(t *testing.T) {
	var _ CreateUserPort
	var _ CreateTransactionUploadPort
	var _ GetAuditEventPort
	var _ DownloadTransactionUploadPort
}
