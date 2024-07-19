export interface FileResult {
  id: string;
  filename: string;
  info: { [key: string]: RowValue };
  toolResults: ToolResults;
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

export interface ToolResults {
  fileIdentificationResults: ToolResult[] | null;
  fileValidationResults: ToolResult[] | null;
  summary: Summary;
}

export interface ToolResult {
  toolName: string;
  toolVersion: string;
  toolOutput: string;
  outputFormat: 'text' | 'json' | 'csv';
  extractedFeatures: { [key: string]: string };
  error: string | null;
}

export interface Summary {
  [key: string]: Feature;
}

export interface Feature {
  key: string;
  values: FeatureValue[];
}

export interface FeatureValue {
  value: string;
  score: number;
  tools: ToolConfidence[];
}

export interface ToolConfidence {
  confidence: number;
  toolName: string;
}
