import { CommonModule } from '@angular/common';
import { Component, Inject } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MAT_DIALOG_DATA, MatDialogModule } from '@angular/material/dialog';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { RouterModule } from '@angular/router';
import { ProcessingError } from './clearing.service';

@Component({
  selector: 'app-clearing-details',
  standalone: true,
  imports: [
    CommonModule,
    MatButtonModule,
    MatDialogModule,
    MatExpansionModule,
    MatFormFieldModule,
    MatIconModule,
    MatInputModule,
    RouterModule,
  ],
  templateUrl: './clearing-details.component.html',
  styleUrl: './clearing-details.component.scss',
})
export class ClearingDetailsComponent {
  json: string;

  constructor(@Inject(MAT_DIALOG_DATA) public data: ProcessingError) {
    this.json = JSON.stringify(data, null, 2);
  }
}
