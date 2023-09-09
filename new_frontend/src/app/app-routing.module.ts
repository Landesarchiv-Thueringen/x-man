import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';

import { MainNavigationComponent } from './main-navigation/main-navigation.component';
import { MessageTableComponent } from './message-table/message-table.component';

const routes: Routes = [
  { 
    path: '',  component: MainNavigationComponent,
    children: [
      { path: 'anbietungen',  component: MessageTableComponent},
      { path: 'abgaben',  component: MessageTableComponent},
    ],
  },
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
