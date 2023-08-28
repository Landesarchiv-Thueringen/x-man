// angular
import { Component } from '@angular/core';

// utility
import * as JSZip from 'jszip';

@Component({
  selector: 'app-main-navigation',
  templateUrl: './main-navigation.component.html',
  styleUrls: ['./main-navigation.component.scss']
})
export class MainNavigationComponent {
  userDisplayName: string;
  messageRegex: RegExp;
  messageText?: string;

  constructor() {
    this.userDisplayName = 'LATh Grochow, Tony';
    this.messageRegex = new RegExp('_Aussonderung\.Anbieteverzeichnis\.0501\.xml$');
  }

  logout(): void {}

  onFileSelected(fileSelectEvent: Event): void {
    const fileList = (fileSelectEvent.target as HTMLInputElement).files;
    if (fileList) {
      const messageContainer = fileList[0];
      this.readXdomeaMessage(messageContainer);
    }
  }

  readXdomeaMessage(messageContainer: File): void {
    if (messageContainer) {
      const zipLib = new JSZip();
      zipLib.loadAsync(messageContainer).then((zip) => {
        const xdomeaMessageFileList = zip.filter((relativePath, zipEntry) => {
          return this.messageRegex.test(relativePath);
        })
        if (xdomeaMessageFileList.length === 1) {
          xdomeaMessageFileList[0].async('text').then((messageText) => {
            this.messageText = messageText;
          })
        } else {
          console.error('multiple xdomea messages in message container');
        }
      });
    } else {
      console.error('no message container selected');
    }
  }

}