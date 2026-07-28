import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';
import { TransferDir } from '../../../services/agencies.service';

export interface TestResult {
  success: boolean;
  connectionSuccess: boolean;
  path0502Exists: boolean;
  path0504Exists: boolean;
  path0506Exists: boolean;
  path0507Exists: boolean;
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
  testTransferDir(transferDir: TransferDir): Observable<TestResult> {
    return this.httpClient.post<TestResult>('/api/test-transfer-dir', transferDir);
  }

  getDefaultRootDir(): string {
    return 'xman/transfer_dir'
  }
}
