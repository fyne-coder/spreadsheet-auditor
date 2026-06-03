package evidence

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"spreadsheet-auditor/internal/model"
)

func BuildPacket(report *model.AuditReport) (*model.EvidencePacketV1, error) {
	auditHash, err := auditHash(report)
	if err != nil {
		return nil, err
	}

	packet := &model.EvidencePacketV1{
		AuditHash:       auditHash,
		FormulaFamilies: []model.FormulaFamily{},
		Issues:          []model.EvidenceIssue{},
		PacketVersion:   model.EvidencePacketVersionV1,
		Sheets:          []model.EvidenceSheet{},
		Workbook: model.EvidenceWorkbook{
			Name:            filepath.Base(report.WorkbookPath),
			SupportedFormat: report.SupportedFormat,
			Summary:         report.ReportSummary(),
		},
	}

	sheetNames := map[string]struct{}{}
	sheetCells := map[string]struct{}{}
	for _, sheet := range report.Sheets {
		packet.Sheets = append(packet.Sheets, model.EvidenceSheet{
			FormulaCells: sheet.FormulaCells,
			Name:         sheet.Name,
			State:        sheet.State,
			UsedRange:    sheet.UsedRange,
		})
		sheetNames[sheet.Name] = struct{}{}
	}

	ruleIDs := map[string]struct{}{}
	issueIDs := map[string]struct{}{}
	clusterIDs := map[string]struct{}{}
	for _, issue := range report.Issues {
		issueID := model.IssueID(issue)
		packet.Issues = append(packet.Issues, model.EvidenceIssue{
			Cell:          issue.Evidence.Cell,
			Category:      issue.Category,
			Details:       sortedDetails(issue.Details),
			Formula:       issue.Evidence.Formula,
			ImpactFactors: evidenceImpactFactors(issue.ImpactFactors),
			IssueID:       issueID,
			Message:       issue.Message,
			Priority:      issue.Priority,
			Rationale:     issue.Rationale,
			Remediation:   issue.Remediation,
			RuleID:        issue.RuleID,
			Severity:      issue.Severity,
			Sheet:         issue.Evidence.Sheet,
			Title:         issue.Title,
		})
		issueIDs[issueID] = struct{}{}
		ruleIDs[issue.RuleID] = struct{}{}
		if issue.Evidence.Sheet != "" && issue.Evidence.Cell != "" {
			sheetCells[sheetCell(issue.Evidence.Sheet, issue.Evidence.Cell)] = struct{}{}
		}
		if family, ok := formulaFamily(issue); ok {
			packet.FormulaFamilies = append(packet.FormulaFamilies, family)
			clusterIDs[family.ClusterID] = struct{}{}
		}
	}

	sort.Slice(packet.FormulaFamilies, func(i, j int) bool {
		return packet.FormulaFamilies[i].ClusterID < packet.FormulaFamilies[j].ClusterID
	})
	packet.CitationMap = model.CitationMap{
		FormulaClusterIDs: sortedSet(clusterIDs),
		IssueIDs:          sortedSet(issueIDs),
		RuleIDs:           sortedSet(ruleIDs),
		SheetCells:        sortedSet(sheetCells),
		SheetNames:        sortedSet(sheetNames),
	}
	return packet, nil
}

func ValidateUnderstanding(packet *model.EvidencePacketV1, report model.UnderstandingReportV1) []model.CitationReject {
	valid := validCitations(packet.CitationMap)
	var rejects []model.CitationReject
	check := func(field string, index int, citations []string) {
		if len(citations) == 0 {
			rejects = append(rejects, model.CitationReject{
				Field:  field,
				Index:  index,
				Reason: "claim does not include any citations",
			})
			return
		}
		for _, citation := range citations {
			if _, ok := valid[citation]; ok {
				continue
			}
			rejects = append(rejects, model.CitationReject{
				Citation: citation,
				Field:    field,
				Index:    index,
				Reason:   "citation is not present in evidence packet citation map",
			})
		}
	}
	for index, claim := range report.WorkbookPurpose {
		check("workbook_purpose.citations", index, claim.Citations)
	}
	for index, claim := range report.SheetRoles {
		check("sheet_roles.citations", index, claim.Citations)
	}
	for index, claim := range report.KeyFlows {
		check("key_flows.citations", index, claim.Citations)
	}
	for index, claim := range report.MajorRisks {
		check("major_risks.citations", index, claim.Citations)
	}
	for index, claim := range report.CleanupPlan {
		check("cleanup_plan.citations", index, claim.Citations)
	}
	for index, claim := range report.OwnerQuestions {
		check("owner_questions.context_citations", index, claim.ContextCitations)
	}
	for index, claim := range report.ConfidenceNotes {
		check("confidence_notes.citations", index, claim.Citations)
	}
	return rejects
}

func ValidateUnderstandingJSON(
	packet *model.EvidencePacketV1,
	raw []byte,
) (*model.UnderstandingReportV1, []model.CitationReject, error) {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, nil, err
	}
	required := []string{
		"workbook_purpose",
		"sheet_roles",
		"key_flows",
		"major_risks",
		"cleanup_plan",
		"owner_questions",
		"confidence_notes",
	}
	for _, key := range required {
		if _, ok := payload[key]; !ok {
			if looksLikeAlternateUnderstandingReport(payload) {
				return nil, nil, fmt.Errorf(
					"assistant response does not match UnderstandingReportV1 root schema: expected top-level sections %s; do not paste wrapper fields such as schema_version, review_objective, workbook_summary, or deterministic_audit_state",
					strings.Join(required, ", "),
				)
			}
			return nil, nil, fmt.Errorf("understanding report missing required section %q", key)
		}
	}
	for key := range payload {
		if !contains(required, key) {
			return nil, nil, fmt.Errorf("understanding report contains unsupported section %q", key)
		}
	}
	var report model.UnderstandingReportV1
	if err := json.Unmarshal(raw, &report); err != nil {
		return nil, nil, err
	}
	return &report, ValidateUnderstanding(packet, report), nil
}

func looksLikeAlternateUnderstandingReport(payload map[string]json.RawMessage) bool {
	alternateKeys := []string{
		"schema_version",
		"review_objective",
		"workbook_summary",
		"deterministic_audit_state",
	}
	for _, key := range alternateKeys {
		if _, ok := payload[key]; ok {
			return true
		}
	}
	return false
}

func auditHash(report *model.AuditReport) (string, error) {
	normalized := *report
	normalized.WorkbookPath = "<workbook>"
	raw, err := normalized.CanonicalJSON()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

func sheetCell(sheet, cell string) string {
	return sheet + "!" + cell
}

func evidenceImpactFactors(factors []model.ImpactFactor) []model.EvidenceImpactFactor {
	if len(factors) == 0 {
		return nil
	}
	result := make([]model.EvidenceImpactFactor, 0, len(factors))
	for _, factor := range factors {
		result = append(result, model.EvidenceImpactFactor{
			Code:        factor.Code,
			Explanation: factor.Explanation,
		})
	}
	return result
}

func validCitations(citations model.CitationMap) map[string]struct{} {
	valid := map[string]struct{}{}
	for _, group := range [][]string{
		citations.FormulaClusterIDs,
		citations.IssueIDs,
		citations.RuleIDs,
		citations.SheetCells,
		citations.SheetNames,
	} {
		for _, value := range group {
			valid[value] = struct{}{}
		}
	}
	return valid
}

func sortedSet(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func sortedDetails(details map[string]any) map[string]any {
	if len(details) == 0 {
		return nil
	}
	allowed := map[string]struct{}{
		"cluster_cells":       {},
		"cluster_orientation": {},
		"constants":           {},
		"expected_pattern":    {},
		"functions":           {},
		"local_pattern":       {},
		"references":          {},
	}
	result := map[string]any{}
	keys := make([]string, 0, len(details))
	for key := range details {
		if _, ok := allowed[key]; ok {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	for _, key := range keys {
		result[key] = details[key]
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func formulaFamily(issue model.Issue) (model.FormulaFamily, bool) {
	if issue.RuleID != "FORMULA_PATTERN_ANOMALY" {
		return model.FormulaFamily{}, false
	}
	clusterCells := stringSliceDetail(issue.Details["cluster_cells"])
	expectedPattern, _ := issue.Details["expected_pattern"].(string)
	localPattern, _ := issue.Details["local_pattern"].(string)
	orientation, _ := issue.Details["cluster_orientation"].(string)
	if len(clusterCells) == 0 || expectedPattern == "" || localPattern == "" {
		return model.FormulaFamily{}, false
	}
	clusterID := stableID("formula_cluster", issue.Evidence.Sheet, strings.Join(clusterCells, ","), expectedPattern, localPattern)
	return model.FormulaFamily{
		ClusterID:       clusterID,
		ExpectedPattern: expectedPattern,
		LocalPattern:    localPattern,
		MemberCells:     clusterCells,
		Orientation:     orientation,
		OutlierCell:     issue.Evidence.Cell,
		Representative:  issue.Evidence.Formula,
		Sheet:           issue.Evidence.Sheet,
	}, true
}

func stringSliceDetail(value any) []string {
	switch items := value.(type) {
	case []string:
		return append([]string(nil), items...)
	case []any:
		result := make([]string, 0, len(items))
		for _, item := range items {
			if text, ok := item.(string); ok {
				result = append(result, text)
			}
		}
		return result
	default:
		return nil
	}
}

func stableID(parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(sum[:])[:16]
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
