// angular
import { Component, ViewChild } from '@angular/core';

// material
import { MatPaginator } from '@angular/material/paginator';
import { MatSort } from '@angular/material/sort';
import { MatTableDataSource } from '@angular/material/table';

// project
import { Message, MessageService } from '../message/message.service';

interface MessageTableRow {
  processID: string;
  agency: string;
}

const DATA: MessageTableRow[] = [
  {processID: '72a1f5a2-cf79-4b9a-8f12-ec649fc3d6b1', agency: 'Arbeitsamt Jena'},
  {processID: 'd2ed87ae-2e85-410c-a4d5-092688d9cd11', agency: 'Thüringer Staatskanzlei'},
]

@Component({
  selector: 'app-message0501-table',
  templateUrl: './message0501-table.component.html',
  styleUrls: ['./message0501-table.component.scss']
})
export class Message0501TableComponent {
  dataSource: MatTableDataSource<Message>;
  displayedColumns: string[] = ['processID', 'agency'];

  @ViewChild(MatPaginator) paginator!: MatPaginator;
  @ViewChild(MatSort) sort!: MatSort;

  constructor(private messageService: MessageService) {
    this.dataSource = new MatTableDataSource();
    this.messageService.get0501Messages().subscribe(
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
}
