import { Component, input } from '@angular/core';
import { MatIconModule } from '@angular/material/icon';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { ItemProgress, TaskState } from '../services/tasks.service';

@Component({
    selector: 'app-task-state-icon',
    imports: [MatIconModule, MatProgressSpinnerModule],
    template: `
    @switch (state()) {
      @case ('pending') {
        <mat-icon>schedule</mat-icon>
      }
      @case ('running') {
        <mat-spinner diameter="24"></mat-spinner>
      }
      @case ('pausing') {
        <mat-spinner diameter="24"></mat-spinner>
      }
      @case ('paused') {
        <mat-spinner
          diameter="24"
          mode="determinate"
          [value]="(progress().done * 100) / progress().total"
        ></mat-spinner>
      }
      @case ('done') {
        <mat-icon class="done">check</mat-icon>
      }
      @case ('failed') {
        <mat-icon class="failed">close</mat-icon>
      }
    }
  `,
    styles: `
    :host {
      display: flex
    }
    .done {
      color: green;
    }
    .failed {
      color: red;
    }
  `
})
export class TaskStateIconComponent {
  readonly state = input.required<TaskState>();
  readonly progress = input.required<ItemProgress>();
}
