import { ComponentFixture, TestBed } from '@angular/core/testing';

import { InstitutMetadataComponent } from './institution-metadata.component';

describe('InstitutMetadataComponent', () => {
  let component: InstitutMetadataComponent;
  let fixture: ComponentFixture<InstitutMetadataComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [InstitutMetadataComponent]
    });
    fixture = TestBed.createComponent(InstitutMetadataComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
