import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { AdminComponent } from './admin/admin.component';
import { AgenciesComponent } from './admin/agencies/agencies.component';
import { CollectionsComponent } from './admin/collections/collections.component';
import { UsersComponent } from './admin/users/users.component';
import { ClearingTableComponent } from './clearing/clearing-table/clearing-table.component';
import { ErrorComponent } from './error/error.component';
import { LoginComponent } from './login/login.component';
import { MainNavigationComponent } from './main-navigation/main-navigation.component';
import { MessageTreeComponent } from './message/message-tree/message-tree.component';
import { DocumentMetadataComponent } from './metadata/document-metadata/document-metadata.component';
import { FileMetadataComponent } from './metadata/file-metadata/file-metadata.component';
import { MessageMetadataComponent } from './metadata/message-metadata/message-metadata.component';
import { PrimaryDocumentsTableComponent } from './metadata/primary-documents-table/primary-documents-table.component';
import { ProcessMetadataComponent } from './metadata/process-metadata/process-metadata.component';
import { ProcessTableComponent } from './process/process-table/process-table.component';
import { isAdmin, isLoggedIn } from './utility/authorization/auth-guards';

const routes: Routes = [
  {
    path: '',
    component: MainNavigationComponent,
    children: [
      { path: 'login', component: LoginComponent },
      {
        path: 'aussonderungen',
        component: ProcessTableComponent,
        canActivate: [isLoggedIn],
      },
      {
        path: 'nachricht/:id',
        component: MessageTreeComponent,
        canActivate: [isLoggedIn],
        children: [
          { path: '', redirectTo: 'details', pathMatch: 'full' },
          { path: 'details', component: MessageMetadataComponent },
          { path: 'akte/:id', component: FileMetadataComponent },
          { path: 'vorgang/:id', component: ProcessMetadataComponent },
          { path: 'dokument/:id', component: DocumentMetadataComponent },
          { path: 'formatverifikation', component: PrimaryDocumentsTableComponent },
        ],
      },
      {
        path: 'steuerungsstelle',
        component: ClearingTableComponent,
        canActivate: [isAdmin],
      },
      {
        path: 'administration',
        component: AdminComponent,
        canActivate: [isAdmin],
        children: [
          { path: '', redirectTo: 'abgebende-stellen', pathMatch: 'full' },
          {
            path: 'abgebende-stellen',
            component: AgenciesComponent,
          },
          {
            path: 'bestände',
            component: CollectionsComponent,
          },
          {
            path: 'mitarbeiter',
            component: UsersComponent,
          },
        ],
      },
      {
        path: 'error/:code',
        component: ErrorComponent,
      },
      { path: '', redirectTo: '/aussonderungen', pathMatch: 'full' },
      { path: '**', redirectTo: '/error/404' },
    ],
  },
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule],
})
export class AppRoutingModule {}
