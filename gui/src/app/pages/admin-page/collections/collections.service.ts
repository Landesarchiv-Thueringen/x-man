import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { BehaviorSubject, Observable } from 'rxjs';
import { filter, first, map, shareReplay } from 'rxjs/operators';
import { environment } from '../../../../environments/environment';
import { Agency } from '../../../services/agencies.service';
import { notNull } from '../../../utils/predicates';

export interface Collection {
  id: number;
  name: string;
  dimagId: string;
}

@Injectable({
  providedIn: 'root',
})
export class CollectionsService {
  private readonly collections = new BehaviorSubject<Collection[] | null>(null);
  private dimagIds?: Observable<string[]>;

  constructor(private httpClient: HttpClient) {
    httpClient
      .get<Collection[]>(environment.endpoint + '/collections')
      .subscribe((collections) => this.collections.next(collections));
  }

  getCollections(): Observable<Collection[]> {
    return this.collections.pipe(first(notNull));
  }

  observeCollections(): Observable<Collection[]> {
    return this.collections.pipe(filter(notNull));
  }

  getCollectionById(id: number): Observable<Collection | null> {
    return this.getCollections().pipe(map((collections) => collections.find((c) => c.id === id) ?? null));
  }

  getAgenciesForCollection(collectionId: number): Observable<Agency[]> {
    return this.httpClient.get<Agency[]>(environment.endpoint + '/agencies', { params: { collectionId } });
  }

  createCollection(collection: Omit<Collection, 'id'>) {
    this.httpClient.put<string>(environment.endpoint + '/collection', collection).subscribe((response) => {
      const id = parseInt(response);
      this.collections.next([...(this.collections.value ?? []), { ...collection, id }]);
    });
  }

  updateCollection(id: number, collection: Omit<Collection, 'id'>) {
    this.httpClient.post<string>(environment.endpoint + '/collection/' + id, collection).subscribe(() => {
      const collections = [...(this.collections.value ?? [])];
      const index = collections.findIndex((c) => c.id === id);
      if (index >= 0) {
        collections[index] = { ...collection, id };
      }
      this.collections.next(collections);
    });
  }

  deleteCollection(collection: Collection) {
    this.collections.next(this.collections.value!.filter((c) => c !== collection));
    this.httpClient.delete(environment.endpoint + '/collection/' + collection.id).subscribe();
  }

  getDimagIds(): Observable<string[]> {
    if (!this.dimagIds) {
      this.dimagIds = this.httpClient.get<string[]>(environment.endpoint + '/collectionDimagIds').pipe(shareReplay(1));
    }
    return this.dimagIds;
  }
}
