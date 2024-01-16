import { ComponentFixture, TestBed } from '@angular/core/testing';

import { ProcessMetadataComponent } from './process-metadata.component';

describe('ProcessMetadataComponent', () => {
  let component: ProcessMetadataComponent;
  let fixture: ComponentFixture<ProcessMetadataComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [ProcessMetadataComponent],
    });
    fixture = TestBed.createComponent(ProcessMetadataComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
