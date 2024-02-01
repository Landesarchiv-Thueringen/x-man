import { AfterViewInit, Component, ViewChild } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { FormControl, ReactiveFormsModule } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatDialog } from '@angular/material/dialog';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatIconModule } from '@angular/material/icon';
import { MatInputModule } from '@angular/material/input';
import { MatSort, MatSortModule } from '@angular/material/sort';
import { MatTableDataSource, MatTableModule } from '@angular/material/table';
import { UserDetailsComponent } from './user-details.component';
import { User, UsersService } from './users.service';

@Component({
  selector: 'app-users',
  standalone: true,
  imports: [
    MatButtonModule,
    MatFormFieldModule,
    MatIconModule,
    MatInputModule,
    MatSortModule,
    MatTableModule,
    ReactiveFormsModule,
  ],
  templateUrl: './users.component.html',
  styleUrl: './users.component.scss',
})
export class UsersComponent implements AfterViewInit {
  @ViewChild(MatSort) sort!: MatSort;

  displayedColumns: string[] = ['icon', 'displayName', 'admin'];
  dataSource = new MatTableDataSource<User>();
  filter = new FormControl('');

  constructor(
    private usersService: UsersService,
    private dialog: MatDialog,
  ) {
    this.usersService
      .getUsers()
      .pipe(takeUntilDestroyed())
      .subscribe((users) => (this.dataSource.data = users));

    this.filter.valueChanges.subscribe((filter) => (this.dataSource.filter = filter as string));
  }

  ngAfterViewInit() {
    this.dataSource.sortingDataAccessor = (item, property) => {
      switch (property) {
        case 'admin':
          return '' + (item.permissions.admin ?? false);
        default:
          return item[property as keyof typeof item] as string;
      }
    };
    this.dataSource.sort = this.sort;
  }

  openDetails(user: User) {
    this.usersService.getAgenciesForUser(user.id).subscribe((agencies) => {
      this.dialog.open(UserDetailsComponent, { data: { user, agencies } });
    });
  }
}
