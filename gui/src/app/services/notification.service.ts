import { Injectable } from '@angular/core';
import { MatSnackBar } from '@angular/material/snack-bar';

@Injectable({
  providedIn: 'root',
})
export class NotificationService {
  messageDuration: number;

  constructor(private snackBar: MatSnackBar) {
    this.messageDuration = 3000;
  }

  show(message: string): void {
    this.snackBar.open(message, undefined, {
      duration: this.messageDuration,
    });
  }
}
