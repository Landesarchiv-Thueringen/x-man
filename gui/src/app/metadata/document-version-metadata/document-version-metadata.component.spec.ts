import { ComponentFixture, TestBed } from '@angular/core/testing';

import { DocumentVersionMetadataComponent } from './document-version-metadata.component';

describe('DocumentVersionMetadataComponent', () => {
  let component: DocumentVersionMetadataComponent;
  let fixture: ComponentFixture<DocumentVersionMetadataComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [DocumentVersionMetadataComponent],
    });
    fixture = TestBed.createComponent(DocumentVersionMetadataComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
