import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable, switchMap } from 'rxjs';
import { environment } from '../../environments/environment';
import { ProcessStep } from './process.service';
import { UpdatesService } from './updates.service';

export type TaskState = 'running' | 'failed' | 'succeeded';
export type TaskType = 'format_verification' | 'archiving';
export interface Task {
  id: string;
  createdAt: string;
  updatedAt: string;
  processId: string;
  processStep: ProcessStep;
  type: TaskType;
  state: TaskState;
  progress: string;
  errorMessage: string;
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
}
