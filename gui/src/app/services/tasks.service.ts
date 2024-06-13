import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable, switchMap } from 'rxjs';
import { environment } from '../../environments/environment';
import { ProcessStep } from './process.service';
import { UpdatesService } from './updates.service';

export interface ItemProgress {
  done: number;
  total: number;
}

export interface TaskItem {
  label: string;
  state: TaskState;
  error: string;
}

export type TaskState = 'pending' | 'running' | 'pausing' | 'paused' | 'failed' | 'done';
export type TaskType = 'format_verification' | 'archiving';
export interface Task {
  id: string;
  createdAt: string;
  updatedAt: string;
  processId: string;
  processStep: ProcessStep;
  type: TaskType;
  state: TaskState;
  progress: ItemProgress;
  error: string;
  items: TaskItem[];
}

@Injectable({
  providedIn: 'root',
})
export class TasksService {
  constructor(
    private httpClient: HttpClient,
    private updates: UpdatesService,
  ) {}

  observeTasks(): Observable<Task[]> {
    return this.updates
      .observeCollection('tasks')
      .pipe(switchMap(() => this.httpClient.get<Task[]>(environment.endpoint + '/tasks')));
  }

  pauseTask(id: string) {
    return this.httpClient.post<void>(environment.endpoint + '/task/action/' + id, 'pause');
  }

  resumeTask(id: string) {
    return this.httpClient.post<void>(environment.endpoint + '/task/action/' + id, 'resume');
  }

  retryTask(id: string) {
    return this.httpClient.post<void>(environment.endpoint + '/task/action/' + id, 'retry');
  }

  cancelTask(id: string) {
    return this.httpClient.post<void>(environment.endpoint + '/task/action/' + id, 'cancel');
  }
}
