import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';

import { MainNavigationComponent } from './main-navigation/main-navigation.component';
import { Message0501TableComponent } from './message0501-table/message0501-table.component';
import { Message0503TableComponent } from './message0503-table/message0503-table.component';
import { MessageViewComponent } from './message-view/message-view.component';
import { MessageMetadataComponent } from './metadata/message-metadata/message-metadata.component';
import { FileMetadataComponent } from './metadata/file-metadata/file-metadata.component';
import { ProcessMetadataComponent } from './metadata/process-metadata/process-metadata.component';

const routes: Routes = [
  { 
    path: '',  component: MainNavigationComponent,
    children: [
      { path: 'anbietungen',  component: Message0501TableComponent },
      { path: 'abgaben',  component: Message0503TableComponent },
      { 
        path: 'nachricht/:id',  component: MessageViewComponent,
        children: [
          { path: 'details', component: MessageMetadataComponent },
          { path: 'akte/:id', component: FileMetadataComponent },
          { path: 'vorgang/:id', component: ProcessMetadataComponent },
        ],
      },
    ],
  },
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
