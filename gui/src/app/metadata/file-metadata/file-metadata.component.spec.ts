import { ComponentFixture, TestBed } from '@angular/core/testing';

import { FileMetadataComponent } from './file-metadata.component';

describe('FileMetadataComponent', () => {
  let component: FileMetadataComponent;
  let fixture: ComponentFixture<FileMetadataComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [FileMetadataComponent],
    });
    fixture = TestBed.createComponent(FileMetadataComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
