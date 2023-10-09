// angular
import { Component } from '@angular/core';

// material
import { MatPaginator } from '@angular/material/paginator';
import { MatSort } from '@angular/material/sort';
import { MatTableDataSource } from '@angular/material/table';

// project
import { Process, ProcessService } from '../process.service';

@Component({
  selector: 'app-process-table',
  templateUrl: './process-table.component.html',
  styleUrls: ['./process-table.component.scss'],
})
export class ProcessTableComponent {
  dataSource: MatTableDataSource<Process>;
  displayedColumns: string[];

  constructor(private processService: ProcessService) {
    this.displayedColumns = ['receivedAt', 'institution'];
    this.dataSource = new MatTableDataSource<Process>();
    this.processService.getProcesses().subscribe({
      error: (error) => {
        console.error(error);
      },
      next: (processes: Process[]) => {
        console.log(processes);
        this.dataSource.data = processes;
      },
    });
  }
}
