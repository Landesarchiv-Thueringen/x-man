import { Injectable } from '@angular/core';
import { FileResult } from './results';

const UNCERTAIN_REQUIRED_FEATURES = ['mimeType', 'puid'];
const UNCERTAIN_CONFIDENCE_THRESHOLD = 0.75;
const VALID_CONFIDENCE_THRESHOLD = 0.75;

export interface StatusIcons {
  uncertain: boolean;
  valid: boolean;
  invalid: boolean;
  error: boolean;
}

@Injectable({
  providedIn: 'root',
})
export class StatusIconsService {
  getIcons(fileResult: FileResult): StatusIcons {
    return {
      uncertain: this.hasUncertainIcon(fileResult),
      valid: this.hasValidIcon(fileResult),
      invalid: this.hasInvalidIcon(fileResult),
      error: this.hasErrorIcon(fileResult),
    };
  }

  private hasUncertainIcon(fileResult: FileResult): boolean {
    for (const key of UNCERTAIN_REQUIRED_FEATURES) {
      const feature = fileResult.toolResults.summary[key];
      if (!feature || feature.values[0].score < UNCERTAIN_CONFIDENCE_THRESHOLD) {
        return true;
      }
    }
    return false;
  }

  private hasValidIcon(fileResult: FileResult): boolean {
    const valid = fileResult.toolResults.summary['valid'];
    return (
      !this.hasUncertainIcon(fileResult) &&
      valid?.values[0].value === 'true' &&
      valid.values[0].score > VALID_CONFIDENCE_THRESHOLD
    );
  }

  private hasInvalidIcon(fileResult: FileResult): boolean {
    const valid = fileResult.toolResults.summary['valid'];
    return (
      !this.hasUncertainIcon(fileResult) &&
      valid?.values[0].value === 'false' &&
      valid.values[0].score > VALID_CONFIDENCE_THRESHOLD
    );
  }

  private hasErrorIcon(fileResult: FileResult): boolean {
    return (
      fileResult.toolResults.fileIdentificationResults?.some((result) => result.error) ||
      fileResult.toolResults.fileValidationResults?.some((result) => result.error) ||
      false
    );
  }
}
