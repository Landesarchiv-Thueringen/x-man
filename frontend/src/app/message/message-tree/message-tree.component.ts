// angular
import { AfterViewInit, Component, OnDestroy } from '@angular/core';
import { ActivatedRoute, Params } from '@angular/router';

// material
import { NestedTreeControl } from '@angular/cdk/tree';
import { MatTreeNestedDataSource } from '@angular/material/tree';

// project
import { Message, MessageService, StructureNode } from '../message.service';
import { NotificationService } from 'src/app/utility/notification/notification.service';

// utility
import { Subscription, switchMap } from 'rxjs';

@Component({
  selector: 'app-message-tree',
  templateUrl: './message-tree.component.html',
  styleUrls: ['./message-tree.component.scss'],
})
export class MessageTreeComponent implements AfterViewInit, OnDestroy {
  treeControl: NestedTreeControl<StructureNode>;
  dataSource: MatTreeNestedDataSource<StructureNode>;
  urlParameterSubscription?: Subscription;
  message?: Message;
  showAppraisal: boolean;

  constructor(
    private messageService: MessageService,
    private notificationService: NotificationService,
    private route: ActivatedRoute
  ) {
    this.showAppraisal = true;
    this.treeControl = new NestedTreeControl<StructureNode>(
      (node) => node.children
    );
    this.dataSource = new MatTreeNestedDataSource<StructureNode>();
  }

  hasChild = (_: number, node: StructureNode) =>
    !!node.children && node.children.length > 0;

  ngAfterViewInit(): void {
    this.urlParameterSubscription?.unsubscribe();
    if (this.route.firstChild) {
      this.urlParameterSubscription = this.route.params
        .pipe(
          switchMap((params: Params) => {
            return this.messageService.getMessage(params['id']);
          }),
          switchMap((message: Message) => {
            this.initTree(message);
            return this.route.firstChild!.params;
          })
        )
        .subscribe((params: Params) => {
          const nodeID: string = params['id'];
          if (nodeID) {
            this.expandNode(nodeID);
          }
        });
    } else {
      this.urlParameterSubscription = this.route.params
        .pipe(
          switchMap((params: Params) => {
            return this.messageService.getMessage(params['id']);
          })
        )
        .subscribe((message: Message) => {
          this.initTree(message);
        });
    }
  }

  initTree(message: Message): void {
    this.message = message;
    this.showAppraisal = this.message.messageType.code === '0501';
    const treeData: StructureNode[] = [];
    const messageNode = this.messageService.processMessage(message);
    treeData.push(messageNode);
    this.dataSource.data = treeData;
    this.treeControl.dataNodes = treeData;
    this.treeControl.expand(messageNode);
  }

  sendAppraisalMessage(): void {
    if (this.message) {
      this.messageService.finalizeMessageAppraisal(this.message.id).subscribe({
        error: (error) => {
          console.error(error);
        },
        next: () => {
          this.notificationService.show('Bewertungsnachricht wurde erfolgreich versandt')
        }
      });
    }
  }

  expandNode(id: string): void {
    const node: StructureNode | undefined =
      this.messageService.getStructureNode(id);
      if (node) {
        this.treeControl.expand(node);
        if (node.parentID) {
          this.expandNode(node.parentID);
        }
      }
  }

  ngOnDestroy(): void {
    this.urlParameterSubscription?.unsubscribe;
  }
}
