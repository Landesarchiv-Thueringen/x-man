// angular
import { Injectable } from '@angular/core';
import { ValidatorFn, Validators } from '@angular/forms';

@Injectable({
  providedIn: 'root',
})
export class ValidationService {
  constructor() {}

  getDateValidator(): ValidatorFn {
    return Validators.pattern(
      '^s*(3[01]|[12][0-9]|0?[1-9]).(1[012]|0?[1-9]).((?:19|20)d{2})s*$'
    );
  }
}
