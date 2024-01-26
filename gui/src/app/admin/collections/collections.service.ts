import { Injectable } from '@angular/core';
import { BehaviorSubject, Observable } from 'rxjs';
import { map } from 'rxjs/operators';

export interface Collection {
  id: string;
  name: string;
}

const dummyCollections: Collection[] = [
  { name: 'Keller', id: '1' },
  { name: 'Flur', id: '2' },
];

@Injectable({
  providedIn: 'root',
})
export class CollectionsService {
  private readonly collections = new BehaviorSubject(dummyCollections);

  constructor() {}

  getCollections(): Observable<Collection[]> {
    return this.collections;
  }

  getCollectionById(id: string): Observable<Collection | null> {
    return this.collections.pipe(map((collections) => collections.find((c) => c.id === id) ?? null));
  }
}
