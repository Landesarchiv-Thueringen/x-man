// angular
import { Component, ViewChild } from '@angular/core';

// material
import { MatPaginator } from '@angular/material/paginator';
import { MatSort } from '@angular/material/sort';
import { MatTableDataSource } from '@angular/material/table';

// project
import { Message, MessageService } from '../message/message.service';

@Component({
  selector: 'app-message0503-table',
  templateUrl: './message0503-table.component.html',
  styleUrls: ['./message0503-table.component.scss']
})
export class Message0503TableComponent {
  dataSource: MatTableDataSource<Message>;
  displayedColumns: string[] = ['creationTime', 'agency', 'processID', 'actions'];

  @ViewChild(MatPaginator) paginator!: MatPaginator;
  @ViewChild(MatSort) sort!: MatSort;

  constructor(private messageService: MessageService) {
    this.dataSource = new MatTableDataSource();
    this.dataSource.sortingDataAccessor = (item: Message, property: string) => {
      switch(property) {
        case 'creationTime':
          return item.messageHead.creationTime;
        case 'agency':
          return item.messageHead.sender.institution.name
        case 'processID':
          return item.messageHead.processID
        default:
          throw new Error('sorting error: unhandled column');
      }
    }
    this.messageService.get0503Messages().subscribe(
      (messages: Message[]) => {
        console.log(messages);
        this.dataSource.data = messages;
      }
    );
    
  }

  ngAfterViewInit() {
    this.dataSource.paginator = this.paginator;
    this.dataSource.sort = this.sort;
  }

  showMessage(messageID: number) {}
}
