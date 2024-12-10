import { Injectable, inject } from '@angular/core';
import { MatSnackBar } from '@angular/material/snack-bar';

@Injectable({
  providedIn: 'root',
})
export class NotificationService {
  private snackBar = inject(MatSnackBar);

  messageDuration: number;

  constructor() {
    this.messageDuration = 3000;
  }

  show(message: string): void {
    this.snackBar.open(message, undefined, {
      duration: this.messageDuration,
    });
  }
}
