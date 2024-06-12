import { CommonModule } from '@angular/common';
import { Component, HostBinding, Inject, effect } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { MatButtonModule } from '@angular/material/button';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatIconModule } from '@angular/material/icon';
import { MatListModule } from '@angular/material/list';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { RouterModule } from '@angular/router';
import { Subject } from 'rxjs';
import { Task, TasksService } from '../../../services/tasks.service';
import { TaskStateIconComponent } from '../../../shared/task-state-icon.component';
import { TaskTitlePipe } from './task-title.pipe';

@Component({
  selector: 'app-task-details',
  standalone: true,
  imports: [
    CommonModule,
    MatButtonModule,
    MatDialogModule,
    MatIconModule,
    MatListModule,
    MatProgressSpinnerModule,
    TaskStateIconComponent,
    TaskTitlePipe,
    RouterModule,
  ],
  templateUrl: './task-details.component.html',
  styleUrl: './task-details.component.scss',
})
export class TaskDetailsComponent {
  task = toSignal(this.taskSubject);
  @HostBinding('class.done') resolved = false;
  @HostBinding('class.failed') failed = false;

  constructor(
    private dialogRef: MatDialogRef<TaskDetailsComponent>,
    @Inject(MAT_DIALOG_DATA) private taskSubject: Subject<Task | undefined>,
    private tasksService: TasksService,
  ) {
    effect(() => {
      this.resolved = this.task()?.state === 'done';
      this.failed = this.task()?.state === 'failed';
      if (this.task() == null) {
        this.dialogRef.close();
      }
    });
  }

  pause(element: Task): void {
    this.tasksService.pauseTask(element.id).subscribe();
  }

  run(element: Task): void {
    this.tasksService.runTask(element.id).subscribe();
  }

  retry(element: Task): void {
    this.tasksService.retryTask(element.id).subscribe();
  }
}
