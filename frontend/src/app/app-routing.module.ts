import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';

//project
import { MainNavigationComponent } from './main-navigation/main-navigation.component';
import { XdomeaFileViewComponent } from './structure/xdomea-file-view/xdomea-file-view.component';

const routes: Routes = [
  { 
    path: 'detail',  component: MainNavigationComponent,
    children: [
      { path: 'akte/:nodeId',  component: XdomeaFileViewComponent},
    ],
  }
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
