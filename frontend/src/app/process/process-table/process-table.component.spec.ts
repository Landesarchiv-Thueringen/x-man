import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ProcessTableComponent } from './process-table.component';

describe('ProcessTableComponent', () => {
  let component: ProcessTableComponent;
  let fixture: ComponentFixture<ProcessTableComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [ProcessTableComponent]
    });
    fixture = TestBed.createComponent(ProcessTableComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
