import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { BehaviorSubject, Observable } from 'rxjs';
import { filter, first, map, shareReplay } from 'rxjs/operators';
import { Agency } from '../../../services/agencies.service';
import { notNull } from '../../../utils/predicates';

export interface ArchiveCollection {
  id: string;
  name: string;
  dimagId: string;
}

@Injectable({
  providedIn: 'root',
})
export class CollectionsService {
  private httpClient = inject(HttpClient);

  private readonly collections = new BehaviorSubject<ArchiveCollection[] | null>(null);
  private dimagIds?: Observable<string[]>;

  constructor() {
    const httpClient = this.httpClient;

    httpClient
      .get<ArchiveCollection[]>('/api/archive-collections')
      .subscribe((collections) => this.collections.next(collections));
  }

  getCollections(): Observable<ArchiveCollection[]> {
    return this.collections.pipe(first(notNull));
  }

  observeCollections(): Observable<ArchiveCollection[]> {
    return this.collections.pipe(filter(notNull));
  }

  getCollectionById(id: string): Observable<ArchiveCollection | null> {
    return this.getCollections().pipe(
      map((collections) => collections.find((c) => c.id === id) ?? null),
    );
  }

  getAgenciesForCollection(collectionId: string): Observable<Agency[]> {
    return this.httpClient.get<Agency[]>('/api/agencies', {
      params: { collectionId },
    });
  }

  createCollection(collection: Omit<ArchiveCollection, 'id'>) {
    this.httpClient
      .put<{ id: string }>('/api/archive-collection', collection)
      .subscribe(({ id }) => {
        this.collections.next([...(this.collections.value ?? []), { ...collection, id }]);
      });
  }

  updateCollection(id: string, collection: Omit<ArchiveCollection, 'id'>) {
    const newCollection = { ...collection, id };
    this.httpClient.post<void>('/api/archive-collection', newCollection).subscribe(() => {
      const collections = [...(this.collections.value ?? [])];
      const index = collections.findIndex((c) => c.id === id);
      if (index >= 0) {
        collections[index] = newCollection;
      }
      this.collections.next(collections);
    });
  }

  deleteCollection(collection: ArchiveCollection) {
    this.collections.next(this.collections.value!.filter((c) => c !== collection));
    this.httpClient.delete('/api/archive-collection/' + collection.id).subscribe();
  }

  getDimagIds(): Observable<string[]> {
    if (!this.dimagIds) {
      this.dimagIds = this.httpClient
        .get<string[]>('/api/dimag-collection-ids')
        .pipe(shareReplay(1));
    }
    return this.dimagIds;
  }
}
