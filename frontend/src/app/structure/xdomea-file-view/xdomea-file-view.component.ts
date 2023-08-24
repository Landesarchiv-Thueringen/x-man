// angular
import { OnInit, Component } from '@angular/core';
import { FormBuilder, FormControl, FormGroup } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';

// project
import { MessageService } from 'src/app/message/message.service';
import { NotificationService } from 'src/app/utility/notification/notification.service';
import { StructureNode } from 'src/app/message/message.service';

@Component({
  selector: 'app-xdomea-file-view',
  templateUrl: './xdomea-file-view.component.html',
  styleUrls: ['./xdomea-file-view.component.scss'],
})
export class XdomeaFileViewComponent implements OnInit {
  form: FormGroup;
  structureNode?: StructureNode;

  constructor(
    private formBuilder: FormBuilder,
    private messageService: MessageService,
    private notificationService: NotificationService,
    private route: ActivatedRoute,
    private router: Router
  ) {
    this.form = this.formBuilder.group({
      recordPlanId: new FormControl<string | null>(''),
      fileId: new FormControl<string | null>(''),
      subject: new FormControl<string | null>(''),
    });
  }

  ngOnInit(): void {
    this.route.params.subscribe((params) => {
      const node = this.messageService.getNode(params['nodeId']);
      if (!node) {
        const errorMessage = 'Akte [' + params['nodeId'] + '] nicht verfügbar';
        this.notificationService.show(errorMessage);
        this.router.navigate(['detail']);
        throw new Error(errorMessage);
      }
      this.structureNode = node;
      const subjectXmlNode = this.messageService.getXmlNodes(
        'xdomea:AllgemeineMetadaten/xdomea:Betreff',
        node.xmlNode
      ).snapshotItem(0);
      const fileIdXmlNode = this.messageService.getXmlNodes(
        'xdomea:AllgemeineMetadaten/xdomea:Kennzeichen',
        node.xmlNode
      ).snapshotItem(0);
      const recordPlanIdXmlNode = this.messageService.getXmlNodes(
        'xdomea:AllgemeineMetadaten/xdomea:Aktenplaneinheit/xdomea:Kennzeichen',
        node.xmlNode,
      ).snapshotItem(0);
      this.form.patchValue({
        recordPlanId: recordPlanIdXmlNode?.textContent,
        fileId: fileIdXmlNode?.textContent,
        subject: subjectXmlNode?.textContent,

      });
    });
  }
}
