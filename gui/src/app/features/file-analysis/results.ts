export interface FileResult {
  id: string;
  filename: string;
  info: { [key: string]: RowValue };
  summary: Summary;
}

export interface RowValue {
  /**
   * The field's value.
   *
   * Used for sorting and display if `displayString` is not set.
   */
  value: string | number;
  /** A string to display instead of `value`. Optional. */
  displayString?: string;
  /** Makes the field a router link to open in a new tab. Optional. */
  routerLink?: any;
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
  puid: string;
  mimeType: string;
  formatVersion: string;
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
