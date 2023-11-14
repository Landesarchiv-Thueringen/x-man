import { Component } from '@angular/core';

@Component({
  selector: 'app-main-navigation',
  templateUrl: './main-navigation.component.html',
  styleUrls: ['./main-navigation.component.scss']
})
export class MainNavigationComponent {
  userDisplayName: string;

  constructor() {
    this.userDisplayName = 'LATh Grochow, Tony';
  }

  logout(): void {}
}
