import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable, interval, startWith, switchMap } from 'rxjs';
import { environment } from '../../../environments/environment';
import { Process, ProcessStep } from '../../process/process.service';

export type TaskState = 'running' | 'failed' | 'succeeded';
export type TaskType = 'formatVerification' | 'archiving';
export interface Task {
  id: number;
  createdAt: string;
  updatedAt: string;
  process: Process;
  processStep: ProcessStep;
  type: TaskType;
  state: TaskState;
  errorMessage: string;
  itemCount: number;
  itemCompletedCount: number;
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
