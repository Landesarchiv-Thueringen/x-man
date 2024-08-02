import { Component } from '@angular/core';
import { MatIconModule, MatIconRegistry } from '@angular/material/icon';
import { DomSanitizer } from '@angular/platform-browser';
import { RouterModule } from '@angular/router';
import { MainNavigationComponent } from './core/main-navigation/main-navigation.component';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss'],
  standalone: true,
  imports: [RouterModule, MainNavigationComponent, MatIconModule],
})
export class AppComponent {
  constructor(
    private matIconRegistry: MatIconRegistry,
    private domSanitizer: DomSanitizer,
  ) {
    this.matIconRegistry.addSvgIcon(
      'folders',
      this.domSanitizer.bypassSecurityTrustResourceUrl('../assets/folders.svg'),
    );
    this.matIconRegistry.addSvgIcon(
      'articles',
      this.domSanitizer.bypassSecurityTrustResourceUrl('../assets/articles.svg'),
    );
    this.matIconRegistry.addSvgIcon(
      'boxes',
      this.domSanitizer.bypassSecurityTrustResourceUrl('../assets/boxes.svg'),
    );
  }
}
