package di

import (
	"testing"

	inboundadapters "github.com/gyud-adb/paris-api/internal/adapters/inbound"
	"github.com/gyud-adb/paris-api/internal/infrastructure/httpserver"
)

// This compile test locks the inbound adapter import path used by DI so the
// adapter split stays wired through the router registrar boundary.
func TestInboundAdaptersImplementRouteRegistrar(t *testing.T) {
	var _ httpserver.RouteRegistrar = (*inboundadapters.HttpAuditEventAdapter)(nil)
	var _ httpserver.RouteRegistrar = (*inboundadapters.HttpBugReportAdapter)(nil)
	var _ httpserver.RouteRegistrar = (*inboundadapters.HttpExclusionListAdapter)(nil)
	var _ httpserver.RouteRegistrar = (*inboundadapters.HttpGroupAdapter)(nil)
	var _ httpserver.RouteRegistrar = (*inboundadapters.HttpSectorAdapter)(nil)
	var _ httpserver.RouteRegistrar = (*inboundadapters.HttpTransactionAdapter)(nil)
	var _ httpserver.RouteRegistrar = (*inboundadapters.HttpTransactionFeedbackAdapter)(nil)
	var _ httpserver.RouteRegistrar = (*inboundadapters.HttpTransactionStep4Adapter)(nil)
	var _ httpserver.RouteRegistrar = (*inboundadapters.HttpTransactionStep5Adapter)(nil)
	var _ httpserver.RouteRegistrar = (*inboundadapters.HttpTransactionUploadAdapter)(nil)
	var _ httpserver.RouteRegistrar = (*inboundadapters.HttpU1ListAdapter)(nil)
	var _ httpserver.RouteRegistrar = (*inboundadapters.HttpUserAdapter)(nil)
}
