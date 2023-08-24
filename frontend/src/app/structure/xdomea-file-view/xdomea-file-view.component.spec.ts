import { ComponentFixture, TestBed } from '@angular/core/testing';

import { XdomeaFileViewComponent } from './xdomea-file-view.component';

describe('XdomeaFileViewComponent', () => {
  let component: XdomeaFileViewComponent;
  let fixture: ComponentFixture<XdomeaFileViewComponent>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      declarations: [XdomeaFileViewComponent]
    });
    fixture = TestBed.createComponent(XdomeaFileViewComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
