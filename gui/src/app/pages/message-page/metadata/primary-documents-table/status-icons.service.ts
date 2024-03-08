import { Injectable } from '@angular/core';
import { PrimaryDocument } from '../../../../services/message.service';

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
  getIcons(primaryDocument: PrimaryDocument): StatusIcons {
    return {
      uncertain: this.hasUncertainIcon(primaryDocument),
      valid: this.hasValidIcon(primaryDocument),
      invalid: this.hasInvalidIcon(primaryDocument),
      error: this.hasErrorIcon(primaryDocument),
    };
  }

  private hasUncertainIcon(primaryDocument: PrimaryDocument): boolean {
    for (const key of UNCERTAIN_REQUIRED_FEATURES) {
      const feature = primaryDocument.formatVerification?.summary[key];
      if (!feature || feature.values[0].score < UNCERTAIN_CONFIDENCE_THRESHOLD) {
        return true;
      }
    }
    return false;
  }

  private hasValidIcon(primaryDocument: PrimaryDocument): boolean {
    const valid = primaryDocument.formatVerification?.summary['valid'];
    return (
      !this.hasUncertainIcon(primaryDocument) &&
      valid?.values[0].value === 'true' &&
      valid.values[0].score > VALID_CONFIDENCE_THRESHOLD
    );
  }

  private hasInvalidIcon(primaryDocument: PrimaryDocument): boolean {
    const valid = primaryDocument.formatVerification?.summary['valid'];
    return (
      !this.hasUncertainIcon(primaryDocument) &&
      valid?.values[0].value === 'false' &&
      valid.values[0].score > VALID_CONFIDENCE_THRESHOLD
    );
  }

  private hasErrorIcon(primaryDocument: PrimaryDocument): boolean {
    return (
      primaryDocument.formatVerification?.fileIdentificationResults?.some((result) => result.error) ||
      primaryDocument.formatVerification?.fileValidationResults?.some((result) => result.error) ||
      false
    );
  }
}
