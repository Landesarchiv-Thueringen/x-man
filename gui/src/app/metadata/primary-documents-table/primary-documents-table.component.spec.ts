import { ComponentFixture, TestBed } from '@angular/core/testing';

import { PrimaryDocumentsTableComponent } from './primary-documents-table.component';

describe('PrimaryDocumentsTableComponent', () => {
  let component: PrimaryDocumentsTableComponent;
  let fixture: ComponentFixture<PrimaryDocumentsTableComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [PrimaryDocumentsTableComponent],
    });
    fixture = TestBed.createComponent(PrimaryDocumentsTableComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
