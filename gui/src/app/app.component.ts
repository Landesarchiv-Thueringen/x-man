import { Component } from '@angular/core';
import { RouterModule } from '@angular/router';
import { MainNavigationComponent } from './core/main-navigation/main-navigation.component';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss'],
  standalone: true,
  imports: [RouterModule, MainNavigationComponent],
})
export class AppComponent {}
