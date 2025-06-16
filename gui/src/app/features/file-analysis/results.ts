export interface FileResult {
  id: string;
  filename: string;
  resourceLink?: ResourceLink;
  summary: Summary;
  additionalMetadata?: { [key: string]: FeatureValue | undefined };
}

export interface ResourceLink {
  sectionLabel: string;
  linkLabel: string;
  routerLink: string[];
}

export interface FileAnalysis {
  summary: Summary;
  featureSets: FeatureSet[];
  toolResults: ToolResult[];
}

export interface Summary {
  valid: boolean;
  invalid: boolean;
  formatUncertain: boolean;
  validityConflict: boolean;
  error: boolean;
  puid: string | null;
  mimeType: string | null;
  formatVersion: string | null;
}

export interface FeatureSet {
  score: number;
  supportingTools: string[];
  features: { [key: string]: FeatureValue | undefined };
}

export interface FeatureValue {
  value: string | boolean | number;
  label: string | null;
  supportingTools: string[];
}

export interface ToolResult {
  id: string;
  title: string;
  toolVersion: string;
  toolOutput: string;
  outputFormat: 'text' | 'json' | 'csv' | 'xml';
  features: { [key: string]: ToolFeatureValue | undefined };
  error: string | null;
}

export interface ToolFeatureValue {
  value: string | boolean | number;
  label: string | null;
}
