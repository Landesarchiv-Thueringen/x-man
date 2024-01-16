import { ComponentFixture, TestBed } from '@angular/core/testing';

import { DocumentMetadataComponent } from './document-metadata.component';

describe('DocumentMetadataComponent', () => {
  let component: DocumentMetadataComponent;
  let fixture: ComponentFixture<DocumentMetadataComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [DocumentMetadataComponent],
    });
    fixture = TestBed.createComponent(DocumentMetadataComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
