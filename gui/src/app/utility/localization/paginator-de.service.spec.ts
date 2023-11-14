import { TestBed } from '@angular/core/testing';

import { PaginatorDeService } from './paginator-de.service';

describe('PaginatorDeService', () => {
  let service: PaginatorDeService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(PaginatorDeService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
