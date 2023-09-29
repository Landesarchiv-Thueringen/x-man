// angular
import { AfterViewInit, Component, OnDestroy, ViewChild } from '@angular/core';

// material
import { MatPaginator } from '@angular/material/paginator';
import { MatSort } from '@angular/material/sort';
import { MatTableDataSource } from '@angular/material/table';

// project
import { Message, MessageService } from '../message.service';

// utility
import { interval, switchMap, Subscription } from 'rxjs';

@Component({
  selector: 'app-message0503-table',
  templateUrl: './message0503-table.component.html',
  styleUrls: ['./message0503-table.component.scss']
})
export class Message0503TableComponent implements AfterViewInit, OnDestroy {
  dataSource: MatTableDataSource<Message>;
  displayedColumns: string[] = ['creationTime', 'agency', 'processID', 'actions'];
  messageSubscription: Subscription;

  @ViewChild(MatPaginator) paginator!: MatPaginator;
  @ViewChild(MatSort) sort!: MatSort;

  constructor(private messageService: MessageService) {
    this.dataSource = new MatTableDataSource();
    this.dataSource.sortingDataAccessor = (item: Message, property: string) => {
      switch(property) {
        case 'creationTime':
          return item.messageHead?.creationTime ? item.messageHead.creationTime : '';
        case 'agency':
          return item.messageHead?.sender?.institution?.name ? item.messageHead.sender.institution.name : ''
        case 'processID':
          return item.messageHead.processID
        default:
          throw new Error('sorting error: unhandled column');
      }
    }
    this.messageService.get0503Messages().subscribe(
      (messages: Message[]) => {
        this.dataSource.data = messages;
      }
    );
    // refetch messages every minute
    // this.messageSubscription = interval(60000)
    this.messageSubscription = interval(5000)
      .pipe(
        switchMap(() => this.messageService.get0503Messages())
      ).subscribe(
        (messages: Message[]) => {
          this.dataSource.data = messages;
        }
      )
  }

  ngAfterViewInit() {
    this.dataSource.paginator = this.paginator;
    this.dataSource.sort = this.sort;
  }

  ngOnDestroy(): void {
    this.messageSubscription.unsubscribe();
  }
}
