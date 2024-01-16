import { Injectable } from '@angular/core';
import { MatSnackBar } from '@angular/material/snack-bar';

@Injectable({
  providedIn: 'root',
})
export class NotificationService {
  messageDuration: number;
  closeSymbol: string;

  constructor(private snackBar: MatSnackBar) {
    this.messageDuration = 3000;
    this.closeSymbol = 'x';
  }

  show(message: string): void {
    this.snackBar.open(message, this.closeSymbol, {
      duration: this.messageDuration,
    });
  }
}
