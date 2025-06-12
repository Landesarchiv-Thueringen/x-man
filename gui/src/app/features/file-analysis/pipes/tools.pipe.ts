import { Pipe, PipeTransform } from '@angular/core';
import { ToolResult } from '../results';

@Pipe({
  name: 'tools',
  standalone: true,
})
export class ToolsPipe implements PipeTransform {
  transform(value: string[], toolResults: ToolResult[]): string {
    const labels = value.map((toolId) => {
      if (toolId === 'browser') {
        return 'Webbrowser';
      }
      const tr = toolResults.find((tr) => tr.id === toolId);
      return tr ? tr.title : toolId;
    });
    return labels.join(', ');
  }
}
