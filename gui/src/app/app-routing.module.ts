import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';

import { ClearingTableComponent } from './clearing/clearing-table/clearing-table.component';
import { DocumentMetadataComponent } from './metadata/document-metadata/document-metadata.component';
import { FileMetadataComponent } from './metadata/file-metadata/file-metadata.component';
import { MainNavigationComponent } from './main-navigation/main-navigation.component';
import { MessageTreeComponent } from './message/message-tree/message-tree.component';
import { MessageMetadataComponent } from './metadata/message-metadata/message-metadata.component';
import { PrimaryDocumentsTableComponent } from './metadata/primary-documents-table/primary-documents-table.component';
import { ProcessMetadataComponent } from './metadata/process-metadata/process-metadata.component';
import { ProcessTableComponent } from './process/process-table/process-table.component';

const routes: Routes = [
  { 
    path: '',  component: MainNavigationComponent,
    children: [
      { path: 'aussonderungen',  component: ProcessTableComponent },
      { 
        path: 'nachricht/:id',  component: MessageTreeComponent,
        children: [
          { path: 'details', component: MessageMetadataComponent },
          { path: 'akte/:id', component: FileMetadataComponent },
          { path: 'vorgang/:id', component: ProcessMetadataComponent },
          { path: 'dokument/:id', component: DocumentMetadataComponent },
          { path: 'formatverifikation', component: PrimaryDocumentsTableComponent },
        ],
      },
      { path: 'steuerungsstelle', component: ClearingTableComponent },
    ],
  },
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
