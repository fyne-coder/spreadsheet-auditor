package main

import (
	"context"
	"fmt"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"spreadsheet-auditor/internal/audit"
	"spreadsheet-auditor/internal/model"
	"spreadsheet-auditor/internal/reviewpack"
)

// AuditService exposes the Go analyzer to the Wails desktop UI.
type AuditService struct {
	ctx context.Context
}

func NewAuditService() *AuditService {
	return &AuditService{}
}

func (s *AuditService) startup(ctx context.Context) {
	s.ctx = ctx
}

func (s *AuditService) ScanWorkbook(path string) (*model.AuditReport, error) {
	return audit.AuditWorkbook(path)
}

func (s *AuditService) RenderReviewPack(path string) (string, error) {
	report, err := audit.AuditWorkbook(path)
	if err != nil {
		return "", err
	}
	return reviewpack.RenderHTML(report, time.Now().UTC()), nil
}

// SaveReviewPack is a deprecated HTML wrapper; prefer SaveExport with format "html".
func (s *AuditService) SaveReviewPack(path string, outputPath string) error {
	return s.SaveExport(path, outputPath, string(reviewpack.FormatHTML), time.Now().UTC().Format(time.RFC3339), nil)
}

// SaveExport scans a workbook and writes HTML or CSV using server-side issue filtering.
func (s *AuditService) SaveExport(
	path string,
	outputPath string,
	format string,
	exportedAtRFC3339 string,
	issueIDs []string,
) error {
	parsedFormat, err := reviewpack.ParseFormat(format)
	if err != nil {
		return err
	}
	exportedAt := time.Now().UTC()
	if exportedAtRFC3339 != "" {
		exportedAt, err = time.Parse(time.RFC3339, exportedAtRFC3339)
		if err != nil {
			return fmt.Errorf("invalid exported-at timestamp: %w", err)
		}
	}
	report, err := audit.AuditWorkbook(path)
	if err != nil {
		return err
	}
	return reviewpack.WriteExport(report, outputPath, reviewpack.ExportOptions{
		Format:     parsedFormat,
		ExportedAt: exportedAt.UTC(),
		IssueIDs:   issueIDs,
	})
}

func (s *AuditService) PickWorkbook() (string, error) {
	if s.ctx == nil {
		return "", nil
	}
	return runtime.OpenFileDialog(s.ctx, runtime.OpenDialogOptions{
		Title: "Select workbook",
		Filters: []runtime.FileFilter{
			{DisplayName: "Excel workbooks (*.xlsx; *.xlsm)", Pattern: "*.xlsx;*.xlsm"},
		},
	})
}

func (s *AuditService) PickReviewPackSavePath(defaultName string) (string, error) {
	return s.PickExportSavePath("html", defaultName)
}

// PickExportSavePath opens a save dialog for HTML or CSV review-pack exports.
func (s *AuditService) PickExportSavePath(format string, defaultName string) (string, error) {
	if s.ctx == nil {
		return "", nil
	}
	parsedFormat, err := reviewpack.ParseFormat(format)
	if err != nil {
		return "", err
	}
	if defaultName == "" {
		switch parsedFormat {
		case reviewpack.FormatCSV:
			defaultName = "review-pack.csv"
		default:
			defaultName = "review-pack.html"
		}
	}
	var filters []runtime.FileFilter
	switch parsedFormat {
	case reviewpack.FormatCSV:
		filters = []runtime.FileFilter{
			{DisplayName: "CSV review pack (*.csv)", Pattern: "*.csv"},
		}
	default:
		filters = []runtime.FileFilter{
			{DisplayName: "HTML review pack (*.html)", Pattern: "*.html"},
		}
	}
	return runtime.SaveFileDialog(s.ctx, runtime.SaveDialogOptions{
		Title:           "Save review pack",
		DefaultFilename: defaultName,
		Filters:         filters,
	})
}
