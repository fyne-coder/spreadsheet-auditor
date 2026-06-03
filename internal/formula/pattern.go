package formula

import (
	"sort"

	"spreadsheet-auditor/internal/model"
)

const minClusterSize = 3

// CellRecord captures one formula cell used for copied-pattern analysis.
type CellRecord struct {
	Coordinate string
	Row        int
	Column     int
	Formula    string
	Pattern    string
}

// FindPatternAnomalies flags a single local outlier inside a copied formula run.
func FindPatternAnomalies(sheet string, records []CellRecord) []model.Issue {
	issues := []model.Issue{}
	flagged := map[string]struct{}{}

	byColumn := map[int][]CellRecord{}
	byRow := map[int][]CellRecord{}
	for _, record := range records {
		byColumn[record.Column] = append(byColumn[record.Column], record)
		byRow[record.Row] = append(byRow[record.Row], record)
	}

	for _, cluster := range columnClusters(byColumn) {
		issue := anomalyIssueForCluster(sheet, cluster)
		if issue == nil {
			continue
		}
		if _, seen := flagged[issue.Evidence.Cell]; seen {
			continue
		}
		flagged[issue.Evidence.Cell] = struct{}{}
		issues = append(issues, *issue)
	}

	for _, cluster := range rowClusters(byRow) {
		issue := anomalyIssueForCluster(sheet, cluster)
		if issue == nil {
			continue
		}
		if _, seen := flagged[issue.Evidence.Cell]; seen {
			continue
		}
		flagged[issue.Evidence.Cell] = struct{}{}
		issues = append(issues, *issue)
	}

	sort.Slice(issues, func(i, j int) bool {
		return issues[i].Evidence.Cell < issues[j].Evidence.Cell
	})
	return issues
}

func columnClusters(recordsByColumn map[int][]CellRecord) [][]CellRecord {
	var clusters [][]CellRecord
	for _, columnRecords := range recordsByColumn {
		ordered := append([]CellRecord(nil), columnRecords...)
		sort.Slice(ordered, func(i, j int) bool {
			return ordered[i].Row < ordered[j].Row
		})
		clusters = append(clusters, splitConsecutiveClusters(ordered, func(record CellRecord) int {
			return record.Row
		})...)
	}
	return clusters
}

func rowClusters(recordsByRow map[int][]CellRecord) [][]CellRecord {
	var clusters [][]CellRecord
	for _, rowRecords := range recordsByRow {
		ordered := append([]CellRecord(nil), rowRecords...)
		sort.Slice(ordered, func(i, j int) bool {
			return ordered[i].Column < ordered[j].Column
		})
		clusters = append(clusters, splitConsecutiveClusters(ordered, func(record CellRecord) int {
			return record.Column
		})...)
	}
	return clusters
}

func splitConsecutiveClusters(ordered []CellRecord, key func(CellRecord) int) [][]CellRecord {
	if len(ordered) == 0 {
		return nil
	}

	var clusters [][]CellRecord
	cluster := []CellRecord{ordered[0]}
	previous := key(ordered[0])
	for _, record := range ordered[1:] {
		current := key(record)
		if current == previous+1 {
			cluster = append(cluster, record)
		} else {
			if len(cluster) >= minClusterSize {
				clusters = append(clusters, cluster)
			}
			cluster = []CellRecord{record}
		}
		previous = current
	}
	if len(cluster) >= minClusterSize {
		clusters = append(clusters, cluster)
	}
	return clusters
}

func anomalyIssueForCluster(sheet string, cluster []CellRecord) *model.Issue {
	patternCounts := map[string]int{}
	for _, record := range cluster {
		patternCounts[record.Pattern]++
	}
	if len(patternCounts) < 2 {
		return nil
	}

	type patternCount struct {
		pattern string
		count   int
	}
	ranked := make([]patternCount, 0, len(patternCounts))
	for pattern, count := range patternCounts {
		ranked = append(ranked, patternCount{pattern: pattern, count: count})
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].count == ranked[j].count {
			return ranked[i].pattern < ranked[j].pattern
		}
		return ranked[i].count > ranked[j].count
	})

	majorityPattern := ranked[0].pattern
	majorityCount := ranked[0].count
	minorityTotal := 0
	for _, entry := range ranked[1:] {
		minorityTotal += entry.count
	}
	if majorityCount != len(cluster)-1 || minorityTotal != 1 {
		return nil
	}

	outlierPattern := ranked[len(ranked)-1].pattern
	var outlier CellRecord
	for _, record := range cluster {
		if record.Pattern == outlierPattern {
			outlier = record
			break
		}
	}

	clusterCoordinates := make([]string, 0, len(cluster))
	for _, record := range cluster {
		clusterCoordinates = append(clusterCoordinates, record.Coordinate)
	}

	orientation := "row"
	columns := map[int]struct{}{}
	for _, record := range cluster {
		columns[record.Column] = struct{}{}
	}
	if len(columns) == 1 {
		orientation = "column"
	}

	issue := model.BuildIssue(
		"FORMULA_PATTERN_ANOMALY",
		"Formula pattern differs from neighboring copied formulas in the same row/column run.",
		sheet,
		outlier.Coordinate,
		outlier.Formula,
		map[string]any{
			"cluster_cells":       clusterCoordinates,
			"cluster_orientation": orientation,
			"expected_pattern":    majorityPattern,
			"local_pattern":       outlierPattern,
		},
	)
	return &issue
}
