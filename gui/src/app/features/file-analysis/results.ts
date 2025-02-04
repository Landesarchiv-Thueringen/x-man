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
  features: { [key: string]: FeatureValue[] };
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

export interface FeatureValue {
  value: string | boolean | number;
  score: number;
  supportingTools: { [key: string]: number };
}

export interface ToolResult {
  toolName: string;
  toolType: 'identification' | 'validation';
  toolVersion: string;
  toolOutput: string;
  outputFormat: 'text' | 'json' | 'csv';
  features: { [key: string]: string };
  error: string | null;
}
