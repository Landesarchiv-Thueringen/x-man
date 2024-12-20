import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';

interface TestResult {
  result: 'success' | 'failed';
}

@Injectable({
  providedIn: 'root',
})
export class TransferDirService {
  private httpClient = inject(HttpClient);

  /**
   * Resolves to `result: 'success'` if the given transfer directory can be
   * reached and is successfully tested for read/write access.
   */
  testTransferDir(transferDir: string): Observable<TestResult> {
    return this.httpClient.post<TestResult>('/api/test-transfer-dir', transferDir);
  }
}
