import { Injectable } from '@angular/core';
import { Observable, map, timer } from 'rxjs';

export interface TransferDirectory {
  uri: string;
  username: string;
  password: string;
}

@Injectable({
  providedIn: 'root',
})
export class TransferDirectoryService {
  constructor() {}

  /**
   * Resolves to `success` if the given transfer directory can be reached and is
   * successfully tested for read/write access.
   */
  testTransferDirectory(transferDirectory: TransferDirectory): Observable<'success' | 'failed'> {
    // TODO: implement
    return timer(1000).pipe(map(() => 'success' as const));
  }
}
