import { Injectable } from '@angular/core';
import { Observable, map, timer } from 'rxjs';

@Injectable({
  providedIn: 'root',
})
export class TransferDirectoryService {
  constructor() {}

  /**
   * Resolves to `success` if the given transfer directory can be reached and is
   * successfully tested for read/write access.
   */
  testTransferDirectory(transferDir: string): Observable<'success' | 'failed'> {
    // TODO: implement
    return timer(1000).pipe(map(() => 'success' as const));
  }
}
