import { BreakpointObserver } from '@angular/cdk/layout';
import { Component, effect, viewChild, inject } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatSidenav, MatSidenavModule } from '@angular/material/sidenav';
import { NavigationEnd, Router, RouterModule } from '@angular/router';
import { filter } from 'rxjs';
import { MessagePageService } from './message-page.service';
import { MessageTreeComponent } from './message-tree/message-tree.component';

@Component({
  selector: 'app-message-page',
  imports: [RouterModule, MatSidenavModule, MatIconModule, MatButtonModule, MessageTreeComponent],
  templateUrl: './message-page.component.html',
  styleUrl: './message-page.component.scss',
  providers: [MessagePageService],
})
export class MessagePageComponent {
  private breakpointObserver = inject(BreakpointObserver);
  private messagePage = inject(MessagePageService);
  private router = inject(Router);

  readonly sidenav = viewChild(MatSidenav);
  sidenavMode: 'side' | 'over' = 'side';

  constructor() {
    // Redirect to the latest message when no message is given in the URL.
    effect(() => {
      if (this.messagePage.messageType() === '') {
        const process = this.messagePage.process();
        if (process) {
          if (process.processState.receive0503.complete) {
            this.router.navigate(['nachricht', process.processId, '0503'], {
              replaceUrl: true,
            });
          } else {
            this.router.navigate(['nachricht', process.processId, '0501'], {
              replaceUrl: true,
            });
          }
        }
      }
    });
    // Redirect to 0503 message when received.
    let message0503NotYetReceived: boolean;
    effect(() => {
      const process = this.messagePage.process();
      if (process && !process.processState.receive0503.complete) {
        message0503NotYetReceived = true;
      } else if (message0503NotYetReceived && process?.processState.receive0503.complete) {
        message0503NotYetReceived = false;
        this.router.navigate(['nachricht', process.processId, '0503']);
      }
    });

    // Show sidenav as overlay on screens smaller than 1700px.
    this.breakpointObserver
      .observe(['(min-width: 1700px)'])
      .pipe(takeUntilDestroyed())
      .subscribe((result) => {
        if (result.matches) {
          this.sidenavMode = 'side';
          this.sidenav()?.open();
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
          this.sidenav()?.close();
        }
      });
  }
}
