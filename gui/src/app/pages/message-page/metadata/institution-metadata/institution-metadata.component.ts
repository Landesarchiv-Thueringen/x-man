import { Component, effect, input, inject } from '@angular/core';
import { FormBuilder, FormControl, ReactiveFormsModule } from '@angular/forms';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { Institution } from '../../../../services/message.service';

@Component({
  selector: 'app-institution-metadata',
  templateUrl: './institution-metadata.component.html',
  styleUrls: ['./institution-metadata.component.scss'],
  imports: [ReactiveFormsModule, MatFormFieldModule, MatInputModule],
})
export class InstitutMetadataComponent {
  private formBuilder = inject(FormBuilder);

  readonly institution = input<Institution>();

  readonly form = this.formBuilder.group({
    abbreviation: new FormControl<string | null>(null),
    name: new FormControl<string | null>(null),
  });

  constructor() {
    effect(() => {
      this.form.patchValue({
        abbreviation: this.institution()?.abbreviation,
        name: this.institution()?.name,
      });
    });
  }
}
