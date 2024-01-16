// angular
import { Component, Input } from '@angular/core';
import { FormBuilder, FormControl, FormGroup } from '@angular/forms';

// project
import { Institution } from '../../message/message.service';

@Component({
  selector: 'app-institution-metadata',
  templateUrl: './institution-metadata.component.html',
  styleUrls: ['./institution-metadata.component.scss'],
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
