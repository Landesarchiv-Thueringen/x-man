import { Component, Input } from '@angular/core';
import { FormBuilder, FormControl, FormGroup, ReactiveFormsModule } from '@angular/forms';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { Institution } from '../../../../services/message.service';

@Component({
  selector: 'app-institution-metadata',
  templateUrl: './institution-metadata.component.html',
  styleUrls: ['./institution-metadata.component.scss'],
  standalone: true,
  imports: [ReactiveFormsModule, MatFormFieldModule, MatInputModule],
})
export class InstitutMetadataComponent {
  form: FormGroup;
  i?: Institution;

  @Input() set institution(i: Institution | null | undefined) {
    if (!!i) {
      this.i = i;
      this.form.patchValue({
        abbrevation: i.abbreviation,
        name: i.name,
      });
    }
  }

  constructor(private formBuilder: FormBuilder) {
    this.form = this.formBuilder.group({
      abbrevation: new FormControl<string | null>(null),
      name: new FormControl<string | null>(null),
    });
  }
}
