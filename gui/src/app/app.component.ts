import { Component, inject } from '@angular/core';
import { MatIconModule, MatIconRegistry } from '@angular/material/icon';
import { DomSanitizer } from '@angular/platform-browser';
import { RouterModule } from '@angular/router';
import { MainNavigationComponent } from './core/main-navigation/main-navigation.component';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss'],
  imports: [RouterModule, MainNavigationComponent, MatIconModule],
})
export class AppComponent {
  private matIconRegistry = inject(MatIconRegistry);
  private domSanitizer = inject(DomSanitizer);

  constructor() {
    this.matIconRegistry.addSvgIcon(
      'folders',
      this.domSanitizer.bypassSecurityTrustResourceUrl('/icons/folders.svg'),
    );
    this.matIconRegistry.addSvgIcon(
      'articles',
      this.domSanitizer.bypassSecurityTrustResourceUrl('/icons/articles.svg'),
    );
    this.matIconRegistry.addSvgIcon(
      'boxes',
      this.domSanitizer.bypassSecurityTrustResourceUrl('/icons/boxes.svg'),
    );
  }
}
