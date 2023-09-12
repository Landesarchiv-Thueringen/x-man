import { Injectable } from '@angular/core';
import { MatPaginatorIntl } from '@angular/material/paginator';

@Injectable({
  providedIn: 'root',
})
export class PaginatorDeService extends MatPaginatorIntl {
  override itemsPerPageLabel = 'Einträge pro Seite';
  override nextPageLabel = 'nächste Seite';
  override previousPageLabel = 'vorherige Seite';
  override firstPageLabel = 'erste Seite';
  override lastPageLabel = 'letzte Seite';
  override getRangeLabel = (
    page: number,
    pageSize: number,
    length: number
  ): string => {
    const preposition = 'von';
    if (length === 0 || pageSize === 0) {
      return '';
    }
    length = Math.max(length, 0);
    const startIndex = page * pageSize;
    // If the start index exceeds the list length, do not try and fix the end index to the end.
    const endIndex =
      startIndex < length
        ? Math.min(startIndex + pageSize, length)
        : startIndex + pageSize;
    return startIndex + 1 + ' - ' + endIndex + ' ' + preposition + ' ' + length;
  };
}