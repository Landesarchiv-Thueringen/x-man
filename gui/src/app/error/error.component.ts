import { Component } from '@angular/core';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatDividerModule } from '@angular/material/divider';
import { ActivatedRoute } from '@angular/router';

/**
 * Error page
 */
@Component({
  selector: 'app-error',
  standalone: true,
  imports: [MatDividerModule, MatDialogModule],
  templateUrl: './error.component.html',
  styleUrl: './error.component.scss',
})
export class ErrorComponent {
  code: string | null = null;

  constructor(route: ActivatedRoute, dialogs: MatDialog) {
    route.params.subscribe((params) => {
      this.code = params['code'];
    });
    dialogs.closeAll();
  }
}
