import { Class2 } from "./index3";

export interface School {
  name: string;
  address: string;
}

export interface School2 {
  name: string;
  location: string;
  class222: Class2
  pack: Package;
}

export type Package = {
  name: string;
  version: string;
}

type Package_Name = {
  name: string;
  version: string;
  location: string;
}

export default Package_Name;