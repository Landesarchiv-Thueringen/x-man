import { Component } from '@angular/core';
import { MatDividerModule } from '@angular/material/divider';
import { ActivatedRoute } from '@angular/router';

/**
 * Error page
 */
@Component({
  selector: 'app-error',
  standalone: true,
  imports: [MatDividerModule],
  templateUrl: './error.component.html',
  styleUrl: './error.component.scss',
})
export class ErrorComponent {
  code: string | null = null;

  constructor(route: ActivatedRoute) {
    route.params.subscribe((params) => {
      this.code = params['code'];
    });
  }
}
