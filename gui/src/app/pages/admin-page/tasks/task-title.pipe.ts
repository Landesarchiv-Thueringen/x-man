import { Pipe, PipeTransform } from '@angular/core';
import { ItemProgress, Task, TaskState, TaskType } from '../../../services/tasks.service';

const titleMap: { [key in TaskType]: string } = {
  archiving: 'Archivierung',
  format_verification: 'Formatverifikation ',
};

@Pipe({
  name: 'taskTitle',
  standalone: true,
})
export class TaskTitlePipe implements PipeTransform {
  transform(task: Task): string {
    const title = titleMap[task.type];
    const progress = getTaskProgressString(task.progress, task.state);
    if (task.progress) {
      return `${title} (${progress})`;
    } else {
      return title;
    }
  }
}

export function getTaskProgressString(progress?: ItemProgress, state?: TaskState): string {
  let result = '';
  if (progress) {
    result = `${progress.done} / ${progress.total}`;
  }
  switch (state) {
    case 'pending':
      result += ' wartet auf Ausführung';
      break;
    case 'paused':
      result += ' pausiert';
      break;
    case 'pausing':
      result += ' wird pausiert...';
      break;
  }
  return result.trim();
}
