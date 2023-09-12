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
      '^(0?[1-9]|[12][0-9]|3[0-1])\\.(0?[1-9]|1[0-2])\\.(\\d{4})$'
    );
  }
}
