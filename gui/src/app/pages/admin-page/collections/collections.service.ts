import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { BehaviorSubject, Observable } from 'rxjs';
import { map } from 'rxjs/operators';
import { environment } from '../../../../environments/environment';
import { Agency } from '../../../services/agencies.service';

export interface Collection {
  id: number;
  name: string;
}

@Injectable({
  providedIn: 'root',
})
export class CollectionsService {
  private readonly collections = new BehaviorSubject<Collection[]>([]);

  constructor(private httpClient: HttpClient) {
    httpClient
      .get<Collection[]>(environment.endpoint + '/collections')
      .subscribe((collections) => this.collections.next(collections));
  }

  getCollections(): Observable<Collection[]> {
    return this.collections;
  }

  getCollectionById(id: number): Observable<Collection | null> {
    return this.collections.pipe(map((collections) => collections.find((c) => c.id === id) ?? null));
  }

  getInstitutionsForCollection(collectionId: number): Observable<Agency[]> {
    return this.httpClient.get<Agency[]>(environment.endpoint + '/agencies', { params: { collectionId } });
  }

  createCollection(collection: Omit<Collection, 'id'>) {
    this.httpClient.put<string>(environment.endpoint + '/collection', collection).subscribe((response) => {
      const id = parseInt(response);
      this.collections.next([...this.collections.value, { ...collection, id }]);
    });
  }

  deleteCollection(collection: Collection) {
    this.collections.next(this.collections.value.filter((c) => c !== collection));
    this.httpClient.delete(environment.endpoint + '/collection/' + collection.id).subscribe();
  }
}
