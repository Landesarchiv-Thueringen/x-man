import { TestBed } from '@angular/core/testing';

import { TransferDirectoryService } from './transfer-directory.service';

describe('TransferDirectoryService', () => {
  let service: TransferDirectoryService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(TransferDirectoryService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
