import { ComponentFixture, TestBed } from '@angular/core/testing';

import { FinalizeAppraisalDialogComponent } from './finalize-appraisal-dialog.component';

describe('FinalizeAppraisalDialogComponent', () => {
  let component: FinalizeAppraisalDialogComponent;
  let fixture: ComponentFixture<FinalizeAppraisalDialogComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [FinalizeAppraisalDialogComponent],
    });
    fixture = TestBed.createComponent(FinalizeAppraisalDialogComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
