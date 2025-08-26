import { OriginalType, RenamedType } from './re-exporter';
import DefaultType from './types';

export interface Container {
  item: OriginalType;
  renamed: RenamedType;
  defaultItem: DefaultType;
}
