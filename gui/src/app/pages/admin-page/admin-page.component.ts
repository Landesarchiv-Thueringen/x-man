import { Component } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { MatButtonModule } from '@angular/material/button';
import { MatDialog } from '@angular/material/dialog';
import { MatIconModule } from '@angular/material/icon';
import { MatListModule } from '@angular/material/list';
import { MatSidenavModule } from '@angular/material/sidenav';
import { RouterModule } from '@angular/router';
import { AboutService } from '../../services/about.service';
import { ConfigService } from '../../services/config.service';
import { AboutDialogComponent } from './about-dialog/about-dialog.component';

@Component({
  selector: 'app-admin-page',
  standalone: true,
  imports: [MatSidenavModule, RouterModule, MatListModule, MatButtonModule, MatIconModule],
  templateUrl: './admin-page.component.html',
  styleUrl: './admin-page.component.scss',
})
export class AdminPageComponent {
  readonly config = this.configService.config;
  readonly aboutInformation = toSignal(this.aboutService.aboutInformation);

  constructor(
    private aboutService: AboutService,
    private configService: ConfigService,
    private dialog: MatDialog,
  ) {}

  openAboutDialog() {
    this.dialog.open(AboutDialogComponent);
  }
}
