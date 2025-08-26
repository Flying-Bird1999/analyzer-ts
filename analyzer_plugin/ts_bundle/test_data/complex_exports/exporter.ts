import { BasicType } from './types';

type MyDefaultType = {
  name: string;
};
export default MyDefaultType;

export interface NamedExportedType {
  child: BasicType;
}

export * from './types';
