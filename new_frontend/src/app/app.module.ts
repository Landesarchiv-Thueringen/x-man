// angular
import { NgModule, LOCALE_ID } from '@angular/core';
import { DatePipe, registerLocaleData } from '@angular/common';
import { BrowserModule } from '@angular/platform-browser';
import { ReactiveFormsModule } from '@angular/forms';
import { HttpClientModule } from '@angular/common/http';
import localeDe from '@angular/common/locales/de';
registerLocaleData(localeDe);

// material
import { MatButtonModule } from '@angular/material/button';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatFormFieldModule } from '@angular/material/form-field'; 
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatMenuModule } from '@angular/material/menu'; 
import { MatPaginatorModule, MatPaginatorIntl} from '@angular/material/paginator';
import { MatSidenavModule } from '@angular/material/sidenav'; 
import { MatSortModule} from '@angular/material/sort';
import { MatTableModule } from '@angular/material/table'; 
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatTreeModule } from '@angular/material/tree';

// project
import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { MessageMetadataComponent } from './metadata/message-metadata/message-metadata.component';
import { FileMetadataComponent } from './metadata/file-metadata/file-metadata.component';
import { InstitutMetadataComponent } from './metadata/institution-metadata/institution-metadata.component';
import { MainNavigationComponent } from './main-navigation/main-navigation.component';
import { Message0501TableComponent } from './message0501-table/message0501-table.component';
import { Message0503TableComponent } from './message0503-table/message0503-table.component';
import { MessageViewComponent } from './message-view/message-view.component';
import { PaginatorDeService } from './utility/localization/paginator-de.service';
import { ProcessMetadataComponent } from './metadata/process-metadata/process-metadata.component';

@NgModule({
  declarations: [
    AppComponent,
    MainNavigationComponent,
    Message0501TableComponent,
    Message0503TableComponent,
    MessageViewComponent,
    MessageMetadataComponent,
    FileMetadataComponent,
    InstitutMetadataComponent,
    ProcessMetadataComponent,
  ],
  imports: [
    BrowserModule,
    AppRoutingModule,
    BrowserAnimationsModule,
    HttpClientModule,
    MatButtonModule,
    MatExpansionModule,
    MatFormFieldModule,
    MatIconModule,
    MatInputModule,
    MatMenuModule,
    MatPaginatorModule,
    MatSidenavModule,
    MatSortModule,
    MatTableModule,
    MatTreeModule,
    MatToolbarModule,
    ReactiveFormsModule,
  ],
  providers: [
    { provide: LOCALE_ID, useValue: 'de' },
    { provide: MatPaginatorIntl, useClass: PaginatorDeService },
    DatePipe,
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }
