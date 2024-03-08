import { Component } from '@angular/core';
import { MatDialog, MatDialogModule } from '@angular/material/dialog';
import { MatDividerModule } from '@angular/material/divider';
import { ActivatedRoute } from '@angular/router';

@Component({
  selector: 'app-error-page',
  standalone: true,
  imports: [MatDividerModule, MatDialogModule],
  templateUrl: './error-page.component.html',
  styleUrl: './error-page.component.scss',
})
export class ErrorPageComponent {
  code: string | null = null;

  constructor(route: ActivatedRoute, dialogs: MatDialog) {
    route.params.subscribe((params) => {
      this.code = params['code'];
    });
    dialogs.closeAll();
  }
}
