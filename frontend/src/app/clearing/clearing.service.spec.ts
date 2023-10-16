import { TestBed } from '@angular/core/testing';

import { ClearingService } from './clearing.service';

describe('ClearingServiceService', () => {
  let service: ClearingService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(ClearingService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
