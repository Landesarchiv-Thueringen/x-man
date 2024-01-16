import { ComponentFixture, TestBed } from '@angular/core/testing';

import { MessageMetadataComponent } from './message-metadata.component';

describe('MessageMetadataComponent', () => {
  let component: MessageMetadataComponent;
  let fixture: ComponentFixture<MessageMetadataComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [MessageMetadataComponent],
    });
    fixture = TestBed.createComponent(MessageMetadataComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
