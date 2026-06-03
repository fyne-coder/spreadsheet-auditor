package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"spreadsheet-auditor/internal/audit"
	"spreadsheet-auditor/internal/evidence"
	"spreadsheet-auditor/internal/model"
	"spreadsheet-auditor/internal/promptpack"
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

// BuildAIHandoff scans once and returns prompt text, canonical JSON payloads,
// and the structured bundle for UI preview. Prefer this for AI handoff panels.
func (s *AuditService) BuildAIHandoff(path string, options model.PromptBundleOptions) (*model.AIHandoffPayload, error) {
	report, err := audit.AuditWorkbook(path)
	if err != nil {
		return nil, err
	}
	return buildAIHandoffPayload(report, options)
}

func (s *AuditService) BuildEvidencePacket(path string) (*model.EvidencePacketV1, error) {
	report, err := audit.AuditWorkbook(path)
	if err != nil {
		return nil, err
	}
	return evidence.BuildPacket(report)
}

func (s *AuditService) BuildPromptBundle(path string, options model.PromptBundleOptions) (*model.PromptBundleV1, error) {
	payload, err := s.BuildAIHandoff(path, options)
	if err != nil {
		return nil, err
	}
	return payload.Bundle, nil
}

func (s *AuditService) BuildEvidencePacketJSON(path string, options model.PromptBundleOptions) (string, error) {
	payload, err := s.BuildAIHandoff(path, options)
	if err != nil {
		return "", err
	}
	return payload.EvidencePacketJSON, nil
}

func (s *AuditService) BuildPromptBundleJSON(path string, options model.PromptBundleOptions) (string, error) {
	payload, err := s.BuildAIHandoff(path, options)
	if err != nil {
		return "", err
	}
	return payload.PromptBundleJSON, nil
}

func (s *AuditService) SaveEvidencePacket(path string, outputPath string, options model.PromptBundleOptions) error {
	payload, err := s.handoffEvidenceJSON(path, options)
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, payload, 0o600)
}

func (s *AuditService) SavePromptBundle(path string, outputPath string, options model.PromptBundleOptions) error {
	payload, err := s.handoffPromptJSON(path, options)
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, payload, 0o600)
}

func (s *AuditService) SaveUnderstandingReport(
	path string,
	rawJSON string,
	outputPath string,
	options model.PromptBundleOptions,
) error {
	result, err := s.ValidateUnderstandingReport(path, rawJSON, options)
	if err != nil {
		return err
	}
	if result.ParseError != "" {
		return fmt.Errorf("cannot save AI analysis: %s", result.ParseError)
	}
	if result.Report == nil || !result.CitationsResolved || len(result.Rejects) > 0 {
		return fmt.Errorf("cannot save AI analysis: cited evidence did not verify")
	}
	payload, err := json.MarshalIndent(result.Report, "", "  ")
	if err != nil {
		return fmt.Errorf("encode AI analysis: %w", err)
	}
	payload = append(payload, '\n')
	return os.WriteFile(outputPath, payload, 0o600)
}

func (s *AuditService) handoffEvidenceJSON(path string, options model.PromptBundleOptions) ([]byte, error) {
	report, err := audit.AuditWorkbook(path)
	if err != nil {
		return nil, err
	}
	payload, err := buildAIHandoffPayload(report, options)
	if err != nil {
		return nil, err
	}
	return []byte(payload.EvidencePacketJSON), nil
}

func (s *AuditService) handoffPromptJSON(path string, options model.PromptBundleOptions) ([]byte, error) {
	report, err := audit.AuditWorkbook(path)
	if err != nil {
		return nil, err
	}
	payload, err := buildAIHandoffPayload(report, options)
	if err != nil {
		return nil, err
	}
	return []byte(payload.PromptBundleJSON), nil
}

func buildAIHandoffPayload(report *model.AuditReport, options model.PromptBundleOptions) (*model.AIHandoffPayload, error) {
	bundle, err := promptpack.BuildBundle(report, options)
	if err != nil {
		return nil, err
	}
	evidenceJSON, err := bundle.EvidencePacket.CanonicalJSON()
	if err != nil {
		return nil, err
	}
	bundleJSON, err := bundle.CanonicalJSON()
	if err != nil {
		return nil, err
	}
	return &model.AIHandoffPayload{
		AuditHash:          bundle.EvidencePacket.AuditHash,
		Prompt:             bundle.Prompt,
		PromptBundleJSON:   string(bundleJSON),
		EvidencePacketJSON: string(evidenceJSON),
		Bundle:             bundle,
	}, nil
}

func (s *AuditService) ValidateUnderstandingReport(
	path string,
	rawJSON string,
	options model.PromptBundleOptions,
) (*model.UnderstandingValidationResult, error) {
	report, err := audit.AuditWorkbook(path)
	if err != nil {
		return nil, err
	}
	packet, err := promptpack.PreparePacket(report, options)
	if err != nil {
		return nil, err
	}
	parsed, rejects, err := evidence.ValidateUnderstandingJSON(packet, []byte(rawJSON))
	result := &model.UnderstandingValidationResult{Rejects: rejects}
	if err != nil {
		result.ParseError = err.Error()
		return result, nil
	}
	result.Report = parsed
	result.CitationsResolved = len(rejects) == 0
	result.Valid = result.CitationsResolved
	return result, nil
}

func (s *AuditService) RenderReviewPack(path string) (string, error) {
	report, err := audit.AuditWorkbook(path)
	if err != nil {
		return "", err
	}
	return reviewpack.RenderHTML(
		report,
		time.Now().UTC(),
		reviewpack.ExportedWorkbookPath(report.WorkbookPath, false),
	), nil
}

// SaveReviewPack is a deprecated HTML wrapper; prefer SaveExport with format "html".
func (s *AuditService) SaveReviewPack(path string, outputPath string) error {
	return s.SaveExport(
		path,
		outputPath,
		string(reviewpack.FormatHTML),
		time.Now().UTC().Format(time.RFC3339),
		nil,
		false,
	)
}

// SaveExport scans a workbook and writes HTML or CSV using server-side issue filtering.
func (s *AuditService) SaveExport(
	path string,
	outputPath string,
	format string,
	exportedAtRFC3339 string,
	issueIDs []string,
	includeFullPath bool,
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
		Format:          parsedFormat,
		ExportedAt:      exportedAt.UTC(),
		IssueIDs:        issueIDs,
		IncludeFullPath: includeFullPath,
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

// PickJSONSavePath opens a save dialog for evidence packet or prompt bundle JSON.
func (s *AuditService) PickJSONSavePath(defaultName string) (string, error) {
	if s.ctx == nil {
		return "", nil
	}
	if defaultName == "" {
		defaultName = "export.json"
	}
	return runtime.SaveFileDialog(s.ctx, runtime.SaveDialogOptions{
		Title:           "Save JSON",
		DefaultFilename: defaultName,
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON (*.json)", Pattern: "*.json"},
		},
	})
}
