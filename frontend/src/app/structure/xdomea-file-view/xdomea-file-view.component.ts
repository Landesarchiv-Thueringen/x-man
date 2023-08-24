// angular
import { AfterViewInit, Component } from '@angular/core';
import { FormBuilder, FormControl, FormGroup } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';

// project
import { MessageService } from 'src/app/message/message.service';
import { NotificationService } from 'src/app/utility/notification/notification.service';

@Component({
  selector: 'app-xdomea-file-view',
  templateUrl: './xdomea-file-view.component.html',
  styleUrls: ['./xdomea-file-view.component.scss'],
})
export class XdomeaFileViewComponent implements AfterViewInit {
  form: FormGroup;

  constructor(
    private formBuilder: FormBuilder,
    private messageService: MessageService,
    private notificationService: NotificationService,
    private route: ActivatedRoute,
    private router: Router,
  ) {
    this.form = this.formBuilder.group({
      subject: new FormControl<string|null>('')
    });
  }

  ngAfterViewInit(): void {
    this.route.params.subscribe((params) => {
      const node = this.messageService.getNode(params['nodeId'])
      if (!node) {
        this.notificationService.show('Akte [' + params['nodeId'] + '] nicht verfügbar');
        this.router.navigate(['detail']);
      }
    })
  }
}
