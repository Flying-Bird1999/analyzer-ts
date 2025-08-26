export interface BasicType {
  id: number;
}

export interface Container {
  p1: MyDefaultType;
  p2: NamedExportedType;
  p3: BasicType;
}

type MyDefaultType = {
  name: string;
};

export interface NamedExportedType {
  child: BasicType;
}
