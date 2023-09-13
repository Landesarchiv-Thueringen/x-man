// angular
import { AfterViewInit, Component } from '@angular/core';
import { DatePipe } from '@angular/common';
import { FormBuilder, FormControl, FormGroup } from '@angular/forms';
import { ActivatedRoute } from '@angular/router';

// project
import { FileRecordObject, MessageService } from '../../message/message.service';

// utility
import { Subscription } from 'rxjs';

@Component({
  selector: 'app-file-metadata',
  templateUrl: './file-metadata.component.html',
  styleUrls: ['./file-metadata.component.scss']
})
export class FileMetadataComponent implements AfterViewInit {
  urlParameterSubscription?: Subscription;
  fileRecordObject?: FileRecordObject;
  form: FormGroup;

  constructor(
    private datePipe: DatePipe,
    private formBuilder: FormBuilder,
    private messageService: MessageService,
    private route: ActivatedRoute,
  ) {
    this.form = this.formBuilder.group({
      recordPlanId: new FormControl<string | null>(null),
      fileId: new FormControl<string | null>(null),
      subject: new FormControl<string | null>(null),
      fileType: new FormControl<string | null>(null),
      lifeStart: new FormControl<string | null>(null),
      lifeEnd: new FormControl<string | null>(null),
      appraisal: new FormControl<number | null>(null),
    });
  }

  ngAfterViewInit(): void {
    this.urlParameterSubscription = this.route.params.subscribe((params) => {
      this.messageService.getFileRecordObject(+params['id']).subscribe(
        (fileRecordObject: FileRecordObject) => {
          this.fileRecordObject = fileRecordObject;
          this.form.patchValue({
            recordPlanId: fileRecordObject.generalMetadata.filePlan.xdomeaID,
            fileId: fileRecordObject.generalMetadata.xdomeaID,
            subject: fileRecordObject.generalMetadata.subject,
            //fileType: fileRecordObject.,
            lifeStart: this.datePipe.transform(new Date(fileRecordObject.lifetime.start)),
            lifeEnd: this.datePipe.transform(new Date(fileRecordObject.lifetime.end)),
          });
        }
      )
    })   
  }
}
