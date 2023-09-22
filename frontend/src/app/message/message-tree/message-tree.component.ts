// angular
import { AfterViewInit, Component, OnDestroy } from '@angular/core';
import { ActivatedRoute, Params } from '@angular/router';

// material
import { NestedTreeControl } from '@angular/cdk/tree';
import { MatTreeNestedDataSource } from '@angular/material/tree';

// project
import { Message, MessageService, StructureNode } from '../message.service';

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

  constructor(
    private messageService: MessageService,
    private route: ActivatedRoute
  ) {
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
    const treeData: StructureNode[] = [];
    const messageNode = this.messageService.processMessage(message);
    treeData.push(messageNode);
    this.dataSource.data = treeData;
    this.treeControl.dataNodes = treeData;
    this.treeControl.expand(messageNode);
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
