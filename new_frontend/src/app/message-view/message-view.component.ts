// angular
import { AfterViewInit, Component } from '@angular/core';
import { ActivatedRoute } from '@angular/router';

// project
import { Message, MessageService } from '../message/message.service';

@Component({
  selector: 'app-message-view',
  templateUrl: './message-view.component.html',
  styleUrls: ['./message-view.component.scss']
})
export class MessageViewComponent implements AfterViewInit{
  constructor(
    private route: ActivatedRoute,
    private messageService: MessageService,
  ) {
    //this.messageService.getMessage
  }

  ngAfterViewInit(): void {
    this.route.params.subscribe((params) => {
      this.messageService.getMessage(+params['id']).subscribe(
        (message: Message) => {
          console.log(message);
        }
      )
    })
  }
}
