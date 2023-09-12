// angular
import { NgModule, LOCALE_ID } from '@angular/core';
import localeDe from '@angular/common/locales/de';
import { registerLocaleData } from '@angular/common';
import { BrowserModule } from '@angular/platform-browser';
import { HttpClientModule } from '@angular/common/http';
registerLocaleData(localeDe);

// project
import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { MainNavigationComponent } from './main-navigation/main-navigation.component';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatMenuModule } from '@angular/material/menu'; 
import { MatPaginatorModule} from '@angular/material/paginator';
import { MatSidenavModule } from '@angular/material/sidenav'; 
import { MatSortModule} from '@angular/material/sort';
import { MatTableModule } from '@angular/material/table'; 
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatTreeModule } from '@angular/material/tree';
import { Message0501TableComponent } from './message0501-table/message0501-table.component';
import { Message0503TableComponent } from './message0503-table/message0503-table.component';
import { MessageViewComponent } from './message-view/message-view.component';
import { MessageMetadataComponent } from './message-metadata/message-metadata.component';
import { FileMetadataComponent } from './file-metadata/file-metadata.component';

@NgModule({
  declarations: [
    AppComponent,
    MainNavigationComponent,
    Message0501TableComponent,
    Message0503TableComponent,
    MessageViewComponent,
    MessageMetadataComponent,
    FileMetadataComponent,
  ],
  imports: [
    BrowserModule,
    AppRoutingModule,
    BrowserAnimationsModule,
    HttpClientModule,
    MatButtonModule,
    MatIconModule,
    MatMenuModule,
    MatPaginatorModule,
    MatSidenavModule,
    MatSortModule,
    MatTableModule,
    MatTreeModule,
    MatToolbarModule,
  ],
  providers: [{ provide: LOCALE_ID, useValue: 'de' }],
  bootstrap: [AppComponent]
})
export class AppModule { }
