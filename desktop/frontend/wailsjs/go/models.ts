export namespace model {
	
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
	export class Issue {
	    RuleID: string;
	    Title: string;
	    Severity: string;
	    Category: string;
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
	        this.RuleID = source["RuleID"];
	        this.Title = source["Title"];
	        this.Severity = source["Severity"];
	        this.Category = source["Category"];
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
	
	
	

}

