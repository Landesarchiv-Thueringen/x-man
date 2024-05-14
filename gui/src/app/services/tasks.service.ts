import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable, interval, startWith, switchMap } from 'rxjs';
import { environment } from '../../environments/environment';
import { ProcessStep } from './process.service';

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
  constructor(private httpClient: HttpClient) {}

  observeTasks(): Observable<Task[]> {
    return interval(environment.updateInterval).pipe(
      startWith(void 0),
      switchMap(() => this.httpClient.get<Task[]>(environment.endpoint + '/tasks')),
    );
  }
}
