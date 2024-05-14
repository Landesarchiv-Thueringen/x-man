import { Pipe, PipeTransform } from '@angular/core';
import { Task, TaskType } from '../../../services/tasks.service';

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
    if (task.progress) {
      return `${title} (${task.progress})`;
    } else {
      return title;
    }
  }
}
