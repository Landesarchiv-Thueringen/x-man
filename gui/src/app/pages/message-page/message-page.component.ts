import { BreakpointObserver } from '@angular/cdk/layout';
import { Component, effect, ViewChild } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatSidenav, MatSidenavModule } from '@angular/material/sidenav';
import { ActivatedRoute, NavigationEnd, Router, RouterModule } from '@angular/router';
import { filter } from 'rxjs';
import { MessagePageService } from './message-page.service';
import { MessageTreeComponent } from './message-tree/message-tree.component';

@Component({
  selector: 'app-message-page',
  standalone: true,
  imports: [RouterModule, MatSidenavModule, MatIconModule, MatButtonModule, MessageTreeComponent],
  templateUrl: './message-page.component.html',
  styleUrl: './message-page.component.scss',
  providers: [MessagePageService],
})
export class MessagePageComponent {
  @ViewChild(MatSidenav) sidenav?: MatSidenav;
  sidenavMode: 'side' | 'over' = 'side';

  constructor(
    route: ActivatedRoute,
    private breakpointObserver: BreakpointObserver,
    private messagePage: MessagePageService,
    private router: Router,
  ) {
    // Redirect to latest message when no message code is given in the URL.
    effect(() => {
      if (this.messagePage.messageType() === '') {
        const process = this.messagePage.process();
        if (process) {
          if (process.processState.receive0503.complete) {
            this.router.navigate(['../0503'], { relativeTo: route, replaceUrl: true });
          } else {
            this.router.navigate(['../0501'], { relativeTo: route, replaceUrl: true });
          }
        }
      }
    });

    // Show sidenav as overlay on screens smaller than 1700px.
    this.breakpointObserver
      .observe(['(min-width: 1700px)'])
      .pipe(takeUntilDestroyed())
      .subscribe((result) => {
        if (result.matches) {
          this.sidenavMode = 'side';
          this.sidenav?.open();
        } else {
          this.sidenavMode = 'over';
        }
      });
    // Close the sidenav on navigation when in overlay mode.
    this.router.events
      .pipe(
        takeUntilDestroyed(),
        filter((e): e is NavigationEnd => e instanceof NavigationEnd),
      )
      .subscribe((event) => {
        if (this.sidenavMode === 'over' && !event.urlAfterRedirects.endsWith('/details')) {
          this.sidenav?.close();
        }
      });
  }
}
