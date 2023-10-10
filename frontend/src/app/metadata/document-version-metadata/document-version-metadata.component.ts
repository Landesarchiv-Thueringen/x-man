// angular
import { Component, Input } from '@angular/core';

import { DocumentVersion } from 'src/app/message/message.service';

@Component({
  selector: 'app-document-version-metadata',
  templateUrl: './document-version-metadata.component.html',
  styleUrls: ['./document-version-metadata.component.scss']
})
export class DocumentVersionMetadataComponent {
  documentVersions?: DocumentVersion[];

  @Input() set versions(v: DocumentVersion[] | null | undefined) {
    if (!!v) {
      this.documentVersions = v;
    }
  }
}
