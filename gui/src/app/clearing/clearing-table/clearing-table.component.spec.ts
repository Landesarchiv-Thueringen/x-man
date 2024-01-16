import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ClearingTableComponent } from './clearing-table.component';

describe('ClearingTableComponent', () => {
  let component: ClearingTableComponent;
  let fixture: ComponentFixture<ClearingTableComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [ClearingTableComponent],
    });
    fixture = TestBed.createComponent(ClearingTableComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
