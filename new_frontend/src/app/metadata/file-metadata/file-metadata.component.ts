// angular
import { AfterViewInit, Component } from '@angular/core';
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
      appraisal: new FormControl<string | null>(null),
      appraisalRecomm: new FormControl<string | null>(null),
    });
  }

  ngAfterViewInit(): void {
    this.urlParameterSubscription = this.route.params.subscribe((params) => {
      this.messageService.getFileRecordObject(+params['id']).subscribe(
        (fileRecordObject: FileRecordObject) => {
          console.log(fileRecordObject);
          this.fileRecordObject = fileRecordObject;
          this.form.patchValue({
            recordPlanId: fileRecordObject.generalMetadata?.filePlan?.xdomeaID,
            fileId: fileRecordObject.generalMetadata?.xdomeaID,
            subject: fileRecordObject.generalMetadata?.subject,
            fileType: fileRecordObject.type,
            lifeStart: this.messageService.getDateText(fileRecordObject.lifetime?.start),
            lifeEnd: this.messageService.getDateText(fileRecordObject.lifetime?.end),
            appraisal: fileRecordObject.archiveMetadata?.appraisalCode,
            appraisalRecomm: fileRecordObject.archiveMetadata?.appraisalRecommCode,
          });
        }
      )
    })   
  }
}
