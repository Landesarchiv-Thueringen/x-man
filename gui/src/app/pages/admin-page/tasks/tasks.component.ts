import { CommonModule } from '@angular/common';
import { AfterViewInit, Component, ViewChild } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSort, MatSortModule } from '@angular/material/sort';
import { MatTableDataSource, MatTableModule } from '@angular/material/table';
import { RouterModule } from '@angular/router';
import { Task, TasksService } from '../../../services/tasks.service';
import { TaskTitlePipe } from './task-title.pipe';

@Component({
  selector: 'app-tasks',
  standalone: true,
  imports: [
    CommonModule,
    MatTableModule,
    MatButtonModule,
    MatIconModule,
    MatProgressSpinnerModule,
    MatSortModule,
    TaskTitlePipe,
    RouterModule,
  ],
  templateUrl: './tasks.component.html',
  styleUrl: './tasks.component.scss',
})
export class TasksComponent implements AfterViewInit {
  @ViewChild(MatSort) sort!: MatSort;

  dataSource = new MatTableDataSource<Task>();
  displayedColumns = ['state', 'type', 'process', 'createdAt', 'updatedAt', 'errorMessage'] as const;

  constructor(private tasksService: TasksService) {
    this.tasksService
      .observeTasks()
      .pipe(takeUntilDestroyed())
      .subscribe((tasks) => (this.dataSource.data = tasks));

    this.dataSource.sortingDataAccessor = ((task: Task, property: TasksComponent['displayedColumns'][number]) => {
      switch (property) {
        case 'process':
          return task.processId;
        default:
          return task[property] ?? '';
      }
    }) as (data: Task, sortHeaderId: string) => string;
  }

  ngAfterViewInit(): void {
    this.dataSource.sort = this.sort;
  }

  trackTableRow(index: number, element: Task): string {
    return element.id;
  }
}
