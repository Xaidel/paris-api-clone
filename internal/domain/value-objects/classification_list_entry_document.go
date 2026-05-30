package valueobjects

import "strings"

// ClassificationListEntryDocument describes one canonical classification list
// entry together with its source row identifier.
type ClassificationListEntryDocument struct {
	sourceRowID string
	entryText   ClassificationListEntryText
}

// NewClassificationListEntryDocument builds a ClassificationListEntryDocument.
func NewClassificationListEntryDocument(sourceRowID string, entryText ClassificationListEntryText) ClassificationListEntryDocument {
	return ClassificationListEntryDocument{
		sourceRowID: strings.TrimSpace(sourceRowID),
		entryText:   entryText,
	}
}

// SourceRowID returns the source row identifier for the canonical entry.
func (d ClassificationListEntryDocument) SourceRowID() string {
	return d.sourceRowID
}

// EntryText returns the canonical entry text.
func (d ClassificationListEntryDocument) EntryText() ClassificationListEntryText {
	return d.entryText
}
