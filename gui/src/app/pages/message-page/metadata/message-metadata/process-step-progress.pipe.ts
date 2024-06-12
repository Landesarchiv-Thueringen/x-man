import { Pipe, PipeTransform } from '@angular/core';
import { ItemProgress, TaskState } from '../../../../services/tasks.service';
import { getTaskProgressString } from '../../../admin-page/tasks/task-title.pipe';

@Pipe({
  name: 'processStepProgressPipe',
  standalone: true,
})
export class ProcessStepProgressPipe implements PipeTransform {
  transform(step: { progress?: ItemProgress; taskState?: TaskState }): string {
    return getTaskProgressString(step.progress, step.taskState);
  }
}
