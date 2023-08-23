import { ComponentFixture, TestBed } from '@angular/core/testing';

import { MessageTreeComponent } from './message-tree.component';

describe('MessageTreeComponent', () => {
  let component: MessageTreeComponent;
  let fixture: ComponentFixture<MessageTreeComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [MessageTreeComponent]
    });
    fixture = TestBed.createComponent(MessageTreeComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
