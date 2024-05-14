import { Routes } from '@angular/router';
import { isAdmin, isLoggedIn } from './core/auth-guards';
import { AdminPageComponent } from './pages/admin-page/admin-page.component';
import { AgenciesComponent } from './pages/admin-page/agencies/agencies.component';
import { CollectionsComponent } from './pages/admin-page/collections/collections.component';
import { TasksComponent } from './pages/admin-page/tasks/tasks.component';
import { UsersComponent } from './pages/admin-page/users/users.component';
import { ClearingPageComponent } from './pages/clearing-page/clearing-page.component';
import { ErrorPageComponent } from './pages/error-page/error-page.component';
import { LoginPageComponent } from './pages/login-page/login-page.component';
import { MessagePageComponent } from './pages/message-page/message-page.component';
import { DocumentMetadataComponent } from './pages/message-page/metadata/document-metadata/document-metadata.component';
import { FileMetadataComponent } from './pages/message-page/metadata/file-metadata/file-metadata.component';
import { MessageMetadataComponent } from './pages/message-page/metadata/message-metadata/message-metadata.component';
import { PrimaryDocumentsTableComponent } from './pages/message-page/metadata/primary-documents-table/primary-documents-table.component';
import { ProcessMetadataComponent } from './pages/message-page/metadata/process-metadata/process-metadata.component';
import { ProcessTablePageComponent } from './pages/process-table-page/process-table-page.component';

export const routes: Routes = [
  { path: 'login', component: LoginPageComponent },
  {
    path: 'aussonderungen',
    component: ProcessTablePageComponent,
    canActivate: [isLoggedIn],
  },
  { path: 'nachricht/:processId', redirectTo: 'nachricht/:processId/' },
  {
    path: 'nachricht/:processId/:messageType',
    component: MessagePageComponent,
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
    component: ClearingPageComponent,
    canActivate: [isAdmin],
  },
  {
    path: 'administration',
    component: AdminPageComponent,
    canActivate: [isAdmin],
    children: [
      { path: '', redirectTo: 'abgebende-stellen', pathMatch: 'full' },
      { path: 'abgebende-stellen', component: AgenciesComponent },
      { path: 'bestände', component: CollectionsComponent },
      { path: 'mitarbeiter', component: UsersComponent },
      { path: 'prozesse', component: TasksComponent },
    ],
  },
  {
    path: 'fehler/:code',
    component: ErrorPageComponent,
  },
  { path: '', redirectTo: '/aussonderungen', pathMatch: 'full' },
  { path: '**', redirectTo: '/fehler/404' },
];
