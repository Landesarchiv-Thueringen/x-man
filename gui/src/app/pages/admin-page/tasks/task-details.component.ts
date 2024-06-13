import { CommonModule } from '@angular/common';
import { Component, HostBinding, Inject, Signal, effect } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { MatButtonModule } from '@angular/material/button';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatIconModule } from '@angular/material/icon';
import { MatListModule } from '@angular/material/list';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { RouterModule } from '@angular/router';
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
  task: Signal<Task | undefined>;
  @HostBinding('class.done') resolved = false;
  @HostBinding('class.failed') failed = false;

  constructor(
    private dialogRef: MatDialogRef<TaskDetailsComponent>,
    @Inject(MAT_DIALOG_DATA) private taskId: string,
    private tasksService: TasksService,
  ) {
    const task = this.tasksService.observeTask(this.taskId);
    this.task = toSignal(task);
    effect(() => {
      this.resolved = this.task()?.state === 'done';
      this.failed = this.task()?.state === 'failed';
    });
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

  cancel(element: Task): void {
    this.tasksService.cancelTask(element.id).subscribe();
  }
}
