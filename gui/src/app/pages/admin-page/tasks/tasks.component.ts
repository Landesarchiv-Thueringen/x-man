import { CommonModule } from '@angular/common';
import { AfterViewInit, Component, viewChild } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { MatButtonModule } from '@angular/material/button';
import { MatDialog } from '@angular/material/dialog';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatSort, MatSortModule } from '@angular/material/sort';
import { MatTableDataSource, MatTableModule } from '@angular/material/table';
import { RouterModule } from '@angular/router';
import { Task, TasksService } from '../../../services/tasks.service';
import { TaskStateIconComponent } from '../../../shared/task-state-icon.component';
import { TaskDetailsComponent } from './task-details.component';
import { TaskTitlePipe } from './task-title.pipe';

@Component({
    selector: 'app-tasks',
    imports: [
        CommonModule,
        MatTableModule,
        MatButtonModule,
        MatIconModule,
        MatProgressSpinnerModule,
        MatSortModule,
        TaskTitlePipe,
        TaskStateIconComponent,
        RouterModule,
    ],
    templateUrl: './tasks.component.html',
    styleUrl: './tasks.component.scss'
})
export class TasksComponent implements AfterViewInit {
  readonly sort = viewChild.required(MatSort);

  dataSource = new MatTableDataSource<Task>();
  displayedColumns = [
    'actions',
    'state',
    'type',
    'process',
    'createdAt',
    'updatedAt',
    'error',
  ] as const;

  constructor(
    private tasksService: TasksService,
    private dialog: MatDialog,
  ) {
    this.tasksService
      .observeTasks()
      .pipe(takeUntilDestroyed())
      .subscribe((tasks) => (this.dataSource.data = tasks));

    this.dataSource.sortingDataAccessor = ((
      task: Task,
      property: TasksComponent['displayedColumns'][number],
    ) => {
      switch (property) {
        case 'process':
          return task.processId;
        case 'actions':
          return null;
        default:
          return task[property] ?? '';
      }
    }) as (data: Task, sortHeaderId: string) => string;
  }

  ngAfterViewInit(): void {
    this.dataSource.sort = this.sort();
  }

  trackTableRow(index: number, element: Task): string {
    return element.id;
  }

  openDetails(task: Task): void {
    this.dialog.open(TaskDetailsComponent, { data: task.id, width: '1000px', maxWidth: '80vw' });
  }

  pause(element: Task): void {
    this.tasksService.pauseTask(element.id).subscribe();
  }

  resume(element: Task): void {
    this.tasksService.resumeTask(element.id).subscribe();
  }

  retry(element: Task): void {
    this.tasksService.retryTask(element.id).subscribe();
  }
}
