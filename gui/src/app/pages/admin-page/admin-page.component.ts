import { Component, inject } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatListModule } from '@angular/material/list';
import { MatSidenavModule } from '@angular/material/sidenav';
import { RouterModule } from '@angular/router';
import { ConfigService } from '../../services/config.service';

@Component({
  selector: 'app-admin-page',
  standalone: true,
  imports: [MatSidenavModule, RouterModule, MatListModule, MatButtonModule, MatIconModule],
  templateUrl: './admin-page.component.html',
  styleUrl: './admin-page.component.scss',
})
export class AdminPageComponent {
  private readonly configService = inject(ConfigService);
  readonly config = toSignal(this.configService.config);
}
