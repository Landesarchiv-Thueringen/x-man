import { AfterViewInit, Component, input, OnInit, QueryList, ViewChildren } from '@angular/core';
import { MatSort, MatSortModule } from '@angular/material/sort';
import { MatTableDataSource, MatTableModule } from '@angular/material/table';
import { BreakOpportunitiesPipe } from '../pipes/break-opportunities.pipe';
import { CategoryPipe } from '../pipes/category.pipe';
import { ToolsPipe } from '../pipes/tools.pipe';
import { FeatureValue, FileAnalysis, FileResult, ToolResult } from '../results';

interface Category {
  id: string;
  features: Feature[];
  dataSource?: MatTableDataSource<Feature>;
}

interface Feature {
  key: string;
  label: string | null;
  value: string | number | boolean;
  supportingTools: string[];
}

@Component({
  selector: 'app-file-metadata',
  imports: [MatTableModule, CategoryPipe, ToolsPipe, MatSortModule, BreakOpportunitiesPipe],
  templateUrl: './file-metadata.component.html',
  styleUrl: './file-metadata.component.scss',
})
export class FileMetadataComponent implements OnInit, AfterViewInit {
  readonly result = input.required<FileResult>();
  readonly fileAnalysis = input.required<FileAnalysis>();
  toolResults: ToolResult[] = [];
  displayedColumns: string[] = ['key', 'value', 'tools'];
  categories: Category[] = [];
  categoryOrder: { [key: string]: number | undefined } = {
    general: 1,
    format: 2,
    av_container: 3,
    audio: 4,
    video: 5,
  };

  @ViewChildren(MatSort) sorts!: QueryList<MatSort>;

  ngOnInit(): void {
    if (this.fileAnalysis().featureSets.length > 0) {
      this.toolResults = this.fileAnalysis().toolResults;
      // merge extracted and additional features
      const features = Object.assign(
        this.fileAnalysis().featureSets[0].features,
        this.result().additionalMetadata,
      );
      let categories = this.getCategories(features);
      this.categories = this.sortCategories(categories);
      for (let category of this.categories) {
        category.dataSource = new MatTableDataSource<Feature>(category.features);
      }
    }
  }

  getCategories(features: { [key: string]: FeatureValue | undefined }): Category[] {
    const categories: Category[] = [];
    for (let key in features) {
      const parts = key.split(':');
      if (parts.length !== 2) {
        console.error('Could not extract category and attribute key from: ' + key);
        continue;
      }
      const categoryKey = parts[0];
      const featureKey = parts[1];
      const feature: Feature = {
        key: featureKey,
        label: features[key]!.label,
        value: features[key]!.value,
        supportingTools: features[key]!.supportingTools,
      };
      const category = categories.find((c) => c.id === categoryKey);
      if (!category) {
        categories.push({
          id: categoryKey,
          features: [feature],
        });
      } else {
        category.features.push(feature);
      }
    }
    return categories;
  }

  ngAfterViewInit(): void {
    const sortArray = this.sorts.toArray();
    sortArray.forEach((sort, index) => {
      this.categories[index].dataSource!.sortingDataAccessor = (item, property) => {
        switch (property) {
          case 'key':
            return item.label ?? item.key;
          case 'value':
            if (typeof item.value === 'boolean') {
              return item.value.toString();
            }
            return item.value;
          case 'tools':
            return item.supportingTools.join('');
          default:
            return '';
        }
      };
      sort.active = 'key';
      sort.direction = 'asc';
      sort.sortChange.emit(); // Triggers the sort logic
      sort._stateChanges.next(); // Triggers the UI update
      this.categories[index].dataSource!.sort = sort;
    });
  }

  sortCategories(categories: Category[]): Category[] {
    return categories.sort((a, b) => {
      const oa = this.categoryOrder[a.id];
      const ob = this.categoryOrder[b.id];
      if (oa && ob) {
        return oa - ob;
      } else if (oa) {
        return -1;
      } else if (ob) {
        return 1;
      }
      return a.id.localeCompare(b.id);
    });
  }
}
