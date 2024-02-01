import { TestBed } from '@angular/core/testing';

import { AgenciesService } from './agencies.service';

describe('AgenciesService', () => {
  let service: AgenciesService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(AgenciesService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
