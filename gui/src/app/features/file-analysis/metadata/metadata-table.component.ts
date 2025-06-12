import { AfterViewInit, Component, input, OnInit, QueryList, ViewChildren } from '@angular/core';
import { MatSort, MatSortModule } from '@angular/material/sort';
import { MatTableDataSource, MatTableModule } from '@angular/material/table';
import { CategoryPipe } from '../pipes/category.pipe';
import { ToolsPipe } from '../pipes/tools.pipe';
import { FileAnalysis, ToolResult } from '../results';

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
  selector: 'app-metadata-table',
  imports: [MatTableModule, CategoryPipe, ToolsPipe, MatSortModule],
  templateUrl: './metadata.component.html',
  styleUrl: './metadata.component.scss',
})
export class MetadataComponent implements OnInit, AfterViewInit {
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
      for (let key in this.fileAnalysis().featureSets[0].features) {
        const parts = key.split(':');
        if (parts.length !== 2) {
          console.error('Could not extract category and attribute key from: ' + key);
          continue;
        }
        const categoryKey = parts[0];
        const featureKey = parts[1];
        const feature: Feature = {
          key: featureKey,
          label: this.fileAnalysis().featureSets[0].features[key]!.label,
          value: this.fileAnalysis().featureSets[0].features[key]!.value,
          supportingTools: this.fileAnalysis().featureSets[0].features[key]!.supportingTools,
        };
        const category = this.categories.find((c) => c.id === categoryKey);
        if (!category) {
          this.categories.push({
            id: categoryKey,
            features: [feature],
          });
        } else {
          category.features.push(feature);
        }
      }
      this.sortCategories();
      for (let category of this.categories) {
        category.dataSource = new MatTableDataSource<Feature>(category.features);
      }
    }
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

  sortCategories(): void {
    this.categories = this.categories.sort((a, b) => {
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
