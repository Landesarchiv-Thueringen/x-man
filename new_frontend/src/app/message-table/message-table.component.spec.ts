import { ComponentFixture, TestBed } from '@angular/core/testing';

import { MessageTableComponent } from './message-table.component';

describe('MessageTableComponent', () => {
  let component: MessageTableComponent;
  let fixture: ComponentFixture<MessageTableComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [MessageTableComponent]
    });
    fixture = TestBed.createComponent(MessageTableComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
