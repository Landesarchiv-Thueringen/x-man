import { ComponentFixture, TestBed } from '@angular/core/testing';

import { StartArchivingDialogComponent } from './start-archiving-dialog.component';

describe('StartArchivingDialogComponent', () => {
  let component: StartArchivingDialogComponent;
  let fixture: ComponentFixture<StartArchivingDialogComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [StartArchivingDialogComponent],
    });
    fixture = TestBed.createComponent(StartArchivingDialogComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
