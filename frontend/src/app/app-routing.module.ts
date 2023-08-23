import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';

//project
import { MainNavigationComponent } from './main-navigation/main-navigation.component';

const routes: Routes = [
  { path: '',  component: MainNavigationComponent}
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
