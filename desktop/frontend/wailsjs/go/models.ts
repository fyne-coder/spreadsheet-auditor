export namespace model {

	export class WorkbookSlice {
	    range: string;
	    sheet: string;
	    values?: string[][];

	    static createFrom(source: any = {}) {
	        return new WorkbookSlice(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.range = source["range"];
	        this.sheet = source["sheet"];
	        this.values = source["values"];
	    }
	}
	export class Summary {
	    SheetCount: number;
	    FormulaCellCount: number;
	    IssueCount: number;
	    IssuesBySeverity: Record<string, number>;
	    IssuesByCategory: Record<string, number>;

	    static createFrom(source: any = {}) {
	        return new Summary(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.SheetCount = source["SheetCount"];
	        this.FormulaCellCount = source["FormulaCellCount"];
	        this.IssueCount = source["IssueCount"];
	        this.IssuesBySeverity = source["IssuesBySeverity"];
	        this.IssuesByCategory = source["IssuesByCategory"];
	    }
	}
	export class EvidenceWorkbook {
	    name: string;
	    supported_format: string;
	    summary: Summary;

	    static createFrom(source: any = {}) {
	        return new EvidenceWorkbook(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.supported_format = source["supported_format"];
	        this.summary = this.convertValues(source["summary"], Summary);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class EvidenceSheet {
	    formula_cells: number;
	    name: string;
	    state: string;
	    used_range: string;

	    static createFrom(source: any = {}) {
	        return new EvidenceSheet(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.formula_cells = source["formula_cells"];
	        this.name = source["name"];
	        this.state = source["state"];
	        this.used_range = source["used_range"];
	    }
	}
	export class EvidenceImpactFactor {
	    code: string;
	    explanation: string;

	    static createFrom(source: any = {}) {
	        return new EvidenceImpactFactor(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.explanation = source["explanation"];
	    }
	}
	export class EvidenceIssue {
	    cell: string;
	    category: string;
	    details?: Record<string, any>;
	    formula?: string;
	    impact_factors?: EvidenceImpactFactor[];
	    issue_id: string;
	    message: string;
	    priority?: string;
	    rationale: string;
	    remediation: string;
	    rule_id: string;
	    severity: string;
	    sheet: string;
	    title: string;

	    static createFrom(source: any = {}) {
	        return new EvidenceIssue(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cell = source["cell"];
	        this.category = source["category"];
	        this.details = source["details"];
	        this.formula = source["formula"];
	        this.impact_factors = this.convertValues(source["impact_factors"], EvidenceImpactFactor);
	        this.issue_id = source["issue_id"];
	        this.message = source["message"];
	        this.priority = source["priority"];
	        this.rationale = source["rationale"];
	        this.remediation = source["remediation"];
	        this.rule_id = source["rule_id"];
	        this.severity = source["severity"];
	        this.sheet = source["sheet"];
	        this.title = source["title"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class FormulaFamily {
	    formula_cluster_id: string;
	    expected_pattern: string;
	    local_pattern: string;
	    member_cells: string[];
	    orientation: string;
	    outlier_cell: string;
	    representative_formula?: string;
	    sheet: string;

	    static createFrom(source: any = {}) {
	        return new FormulaFamily(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.formula_cluster_id = source["formula_cluster_id"];
	        this.expected_pattern = source["expected_pattern"];
	        this.local_pattern = source["local_pattern"];
	        this.member_cells = source["member_cells"];
	        this.orientation = source["orientation"];
	        this.outlier_cell = source["outlier_cell"];
	        this.representative_formula = source["representative_formula"];
	        this.sheet = source["sheet"];
	    }
	}
	export class CitationMap {
	    formula_cluster_ids: string[];
	    issue_ids: string[];
	    rule_ids: string[];
	    sheet_cells: string[];
	    sheet_names: string[];

	    static createFrom(source: any = {}) {
	        return new CitationMap(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.formula_cluster_ids = source["formula_cluster_ids"];
	        this.issue_ids = source["issue_ids"];
	        this.rule_ids = source["rule_ids"];
	        this.sheet_cells = source["sheet_cells"];
	        this.sheet_names = source["sheet_names"];
	    }
	}
	export class EvidencePacketV1 {
	    audit_hash: string;
	    citation_map: CitationMap;
	    formula_families: FormulaFamily[];
	    audit_findings: EvidenceIssue[];
	    packet_version: string;
	    sheets: EvidenceSheet[];
	    workbook: EvidenceWorkbook;

	    static createFrom(source: any = {}) {
	        return new EvidencePacketV1(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.audit_hash = source["audit_hash"];
	        this.citation_map = this.convertValues(source["citation_map"], CitationMap);
	        this.formula_families = this.convertValues(source["formula_families"], FormulaFamily);
	        this.audit_findings = this.convertValues(source["audit_findings"], EvidenceIssue);
	        this.packet_version = source["packet_version"];
	        this.sheets = this.convertValues(source["sheets"], EvidenceSheet);
	        this.workbook = this.convertValues(source["workbook"], EvidenceWorkbook);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class PromptBundleV1 {
	    bundle_version: string;
	    prompt_version: string;
	    instructions: string;
	    response_schema: Record<string, any>;
	    evidence_packet: EvidencePacketV1;
	    workbook_slices?: WorkbookSlice[];
	    prompt: string;

	    static createFrom(source: any = {}) {
	        return new PromptBundleV1(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.bundle_version = source["bundle_version"];
	        this.prompt_version = source["prompt_version"];
	        this.instructions = source["instructions"];
	        this.response_schema = source["response_schema"];
	        this.evidence_packet = this.convertValues(source["evidence_packet"], EvidencePacketV1);
	        this.workbook_slices = this.convertValues(source["workbook_slices"], WorkbookSlice);
	        this.prompt = source["prompt"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class AIHandoffPayload {
	    audit_hash: string;
	    prompt: string;
	    prompt_bundle_json: string;
	    evidence_packet_json: string;
	    bundle?: PromptBundleV1;

	    static createFrom(source: any = {}) {
	        return new AIHandoffPayload(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.audit_hash = source["audit_hash"];
	        this.prompt = source["prompt"];
	        this.prompt_bundle_json = source["prompt_bundle_json"];
	        this.evidence_packet_json = source["evidence_packet_json"];
	        this.bundle = this.convertValues(source["bundle"], PromptBundleV1);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class IssueEvidence {
	    Sheet: string;
	    Cell: string;
	    Formula: string;

	    static createFrom(source: any = {}) {
	        return new IssueEvidence(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Sheet = source["Sheet"];
	        this.Cell = source["Cell"];
	        this.Formula = source["Formula"];
	    }
	}
	export class ImpactFactor {
	    Code: string;
	    Explanation: string;

	    static createFrom(source: any = {}) {
	        return new ImpactFactor(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Code = source["Code"];
	        this.Explanation = source["Explanation"];
	    }
	}
	export class Issue {
	    IssueID: string;
	    RuleID: string;
	    Title: string;
	    Severity: string;
	    Category: string;
	    Priority: string;
	    ImpactFactors: ImpactFactor[];
	    Rationale: string;
	    Remediation: string;
	    Message: string;
	    Evidence: IssueEvidence;
	    Details: Record<string, any>;

	    static createFrom(source: any = {}) {
	        return new Issue(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.IssueID = source["IssueID"];
	        this.RuleID = source["RuleID"];
	        this.Title = source["Title"];
	        this.Severity = source["Severity"];
	        this.Category = source["Category"];
	        this.Priority = source["Priority"];
	        this.ImpactFactors = this.convertValues(source["ImpactFactors"], ImpactFactor);
	        this.Rationale = source["Rationale"];
	        this.Remediation = source["Remediation"];
	        this.Message = source["Message"];
	        this.Evidence = this.convertValues(source["Evidence"], IssueEvidence);
	        this.Details = source["Details"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SheetSummary {
	    Name: string;
	    State: string;
	    UsedRange: string;
	    FormulaCells: number;

	    static createFrom(source: any = {}) {
	        return new SheetSummary(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.State = source["State"];
	        this.UsedRange = source["UsedRange"];
	        this.FormulaCells = source["FormulaCells"];
	    }
	}
	export class AuditReport {
	    WorkbookPath: string;
	    SupportedFormat: string;
	    Summary: Summary;
	    Sheets: SheetSummary[];
	    Issues: Issue[];

	    static createFrom(source: any = {}) {
	        return new AuditReport(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.WorkbookPath = source["WorkbookPath"];
	        this.SupportedFormat = source["SupportedFormat"];
	        this.Summary = this.convertValues(source["Summary"], Summary);
	        this.Sheets = this.convertValues(source["Sheets"], SheetSummary);
	        this.Issues = this.convertValues(source["Issues"], Issue);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

	export class CitationReject {
	    citation: string;
	    field: string;
	    index: number;
	    reason: string;

	    static createFrom(source: any = {}) {
	        return new CitationReject(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.citation = source["citation"];
	        this.field = source["field"];
	        this.index = source["index"];
	        this.reason = source["reason"];
	    }
	}
	export class CleanupAction {
	    action: string;
	    citations: string[];

	    static createFrom(source: any = {}) {
	        return new CleanupAction(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.action = source["action"];
	        this.citations = source["citations"];
	    }
	}
	export class ConfidenceNote {
	    citations: string[];
	    note: string;

	    static createFrom(source: any = {}) {
	        return new ConfidenceNote(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.citations = source["citations"];
	        this.note = source["note"];
	    }
	}





	export class FlowClaim {
	    citations: string[];
	    summary: string;

	    static createFrom(source: any = {}) {
	        return new FlowClaim(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.citations = source["citations"];
	        this.summary = source["summary"];
	    }
	}




	export class OwnerQuestion {
	    context_citations: string[];
	    question: string;

	    static createFrom(source: any = {}) {
	        return new OwnerQuestion(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.context_citations = source["context_citations"];
	        this.question = source["question"];
	    }
	}
	export class PromptBundleOptions {
	    exclude_sheets?: string[];
	    exclude_cells?: string[];
	    enable_workbook_slices?: boolean;
	    max_slice_rows?: number;
	    max_slice_columns?: number;
	    max_packet_bytes?: number;
	    user_objective?: string;

	    static createFrom(source: any = {}) {
	        return new PromptBundleOptions(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.exclude_sheets = source["exclude_sheets"];
	        this.exclude_cells = source["exclude_cells"];
	        this.enable_workbook_slices = source["enable_workbook_slices"];
	        this.max_slice_rows = source["max_slice_rows"];
	        this.max_slice_columns = source["max_slice_columns"];
	        this.max_packet_bytes = source["max_packet_bytes"];
	        this.user_objective = source["user_objective"];
	    }
	}

	export class RiskClaim {
	    citations: string[];
	    severity: string;
	    summary: string;

	    static createFrom(source: any = {}) {
	        return new RiskClaim(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.citations = source["citations"];
	        this.severity = source["severity"];
	        this.summary = source["summary"];
	    }
	}
	export class SheetRoleClaim {
	    citations: string[];
	    role: string;
	    sheet: string;

	    static createFrom(source: any = {}) {
	        return new SheetRoleClaim(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.citations = source["citations"];
	        this.role = source["role"];
	        this.sheet = source["sheet"];
	    }
	}


	export class UnderstandingClaim {
	    citations: string[];
	    claim: string;

	    static createFrom(source: any = {}) {
	        return new UnderstandingClaim(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.citations = source["citations"];
	        this.claim = source["claim"];
	    }
	}
	export class UnderstandingReportV1 {
	    cleanup_plan: CleanupAction[];
	    confidence_notes: ConfidenceNote[];
	    key_flows: FlowClaim[];
	    major_risks: RiskClaim[];
	    owner_questions: OwnerQuestion[];
	    sheet_roles: SheetRoleClaim[];
	    workbook_purpose: UnderstandingClaim[];

	    static createFrom(source: any = {}) {
	        return new UnderstandingReportV1(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cleanup_plan = this.convertValues(source["cleanup_plan"], CleanupAction);
	        this.confidence_notes = this.convertValues(source["confidence_notes"], ConfidenceNote);
	        this.key_flows = this.convertValues(source["key_flows"], FlowClaim);
	        this.major_risks = this.convertValues(source["major_risks"], RiskClaim);
	        this.owner_questions = this.convertValues(source["owner_questions"], OwnerQuestion);
	        this.sheet_roles = this.convertValues(source["sheet_roles"], SheetRoleClaim);
	        this.workbook_purpose = this.convertValues(source["workbook_purpose"], UnderstandingClaim);
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class UnderstandingValidationResult {
	    citations_resolved: boolean;
	    valid: boolean;
	    report?: UnderstandingReportV1;
	    rejects?: CitationReject[];
	    parse_error?: string;

	    static createFrom(source: any = {}) {
	        return new UnderstandingValidationResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.citations_resolved = source["citations_resolved"];
	        this.valid = source["valid"];
	        this.report = this.convertValues(source["report"], UnderstandingReportV1);
	        this.rejects = this.convertValues(source["rejects"], CitationReject);
	        this.parse_error = source["parse_error"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}
