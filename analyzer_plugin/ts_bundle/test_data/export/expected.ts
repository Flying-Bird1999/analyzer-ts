export interface Container {
  item: RenamedType;
  renamed: RenamedType;
  defaultItem: DefaultType;
}

export type DefaultType = {
    message: string;
}

export interface RenamedType {
  id: number;
}