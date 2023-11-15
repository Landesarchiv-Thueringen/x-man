// angular
import { AfterViewInit, Component } from '@angular/core';
import { ActivatedRoute, Params } from '@angular/router';

// project
import { MessageService, PrimaryDocument } from '../message.service';

// utility
import { Subscription, switchMap } from 'rxjs';

@Component({
  selector: 'app-primary-documents-table',
  templateUrl: './primary-documents-table.component.html',
  styleUrls: ['./primary-documents-table.component.scss'],
})
export class PrimaryDocumentsTableComponent implements AfterViewInit {
  urlParameterSubscription?: Subscription;
  constructor(
    private messageService: MessageService,
    private route: ActivatedRoute
  ) {}

  ngAfterViewInit(): void {
    const messageID: string = this.route.parent!.snapshot.params['id'];
    this.messageService.getPrimaryDocuments(messageID).subscribe({
      error: (error: any) => {
        console.error(error);
      },
      next: (primaryDocuments: PrimaryDocument[]) => {
        console.log(primaryDocuments);
      },
    });
  }
}
