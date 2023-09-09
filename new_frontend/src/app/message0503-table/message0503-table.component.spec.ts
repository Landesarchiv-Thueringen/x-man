import { ComponentFixture, TestBed } from '@angular/core/testing';

import { Message0503TableComponent } from './message0503-table.component';

describe('Message0503TableComponent', () => {
  let component: Message0503TableComponent;
  let fixture: ComponentFixture<Message0503TableComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [Message0503TableComponent]
    });
    fixture = TestBed.createComponent(Message0503TableComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
