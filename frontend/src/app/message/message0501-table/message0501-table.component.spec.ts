import { ComponentFixture, TestBed } from '@angular/core/testing';

import { Message0501TableComponent } from './message0501-table.component';

describe('MessageTableComponent', () => {
  let component: Message0501TableComponent;
  let fixture: ComponentFixture<Message0501TableComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [Message0501TableComponent]
    });
    fixture = TestBed.createComponent(Message0501TableComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
