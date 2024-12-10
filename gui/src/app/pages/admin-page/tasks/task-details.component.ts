import { ScrollingModule } from '@angular/cdk/scrolling';
import { CommonModule } from '@angular/common';
import { Component, HostBinding, Signal, effect, inject } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { MatButtonModule } from '@angular/material/button';
import { MAT_DIALOG_DATA, MatDialogModule, MatDialogRef } from '@angular/material/dialog';
import { MatIconModule } from '@angular/material/icon';
import { MatListModule } from '@angular/material/list';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { RouterModule } from '@angular/router';
import { AuthService } from '../../../services/auth.service';
import { Task, TasksService } from '../../../services/tasks.service';
import { TaskTitlePipe } from './task-title.pipe';

@Component({
  selector: 'app-task-details',
  imports: [
    CommonModule,
    MatButtonModule,
    MatDialogModule,
    MatIconModule,
    MatListModule,
    MatProgressSpinnerModule,
    RouterModule,
    ScrollingModule,
    TaskTitlePipe,
  ],
  templateUrl: './task-details.component.html',
  styleUrl: './task-details.component.scss',
})
export class TaskDetailsComponent {
  private dialogRef = inject<MatDialogRef<TaskDetailsComponent>>(MatDialogRef);
  private taskId = inject(MAT_DIALOG_DATA);
  private auth = inject(AuthService);
  private tasksService = inject(TasksService);

  @HostBinding('class.done') resolved = false;
  @HostBinding('class.failed') failed = false;
  task: Signal<Task | undefined>;
  isAdmin = this.auth.isAdmin();

  constructor() {
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
